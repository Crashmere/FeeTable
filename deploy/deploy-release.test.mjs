import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, readFileSync, writeFileSync, chmodSync, existsSync, readdirSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawnSync } from 'node:child_process';
import { createHash } from 'node:crypto';

// 在临时目录执行生产脚本的测试副本，systemctl/curl/runuser 全部替换为合成命令。
// 不连接服务器，不读取正式账本；生产脚本本身不提供路径或权限绕过开关。
function fixture() {
  const root = mkdtempSync(join(tmpdir(), 'feetable-deploy-test.'));
  const app = join(root, 'app');
  const mocks = join(root, 'mocks');
  for (const dir of [mocks, app, join(app, 'bin'), join(app, 'data'), join(app, 'backups')]) mkdirSync(dir);
  writeFileSync(join(app, 'bin/feetable'), 'old-program');
  chmodSync(join(app, 'bin/feetable'), 0o755);
  writeFileSync(join(app, 'data/feetable.sqlite'), 'synthetic-freight-data');
  writeFileSync(join(root, 'state'), 'active');
  const command = (name, script) => {
    const path = join(mocks, name);
    writeFileSync(path, '#!/bin/bash\nset -e\n' + script + '\n');
    chmodSync(path, 0o755);
  };
  command('flock', 'exit 0');
  command('timeout', 'shift; exec "$@"');
  command('sleep', 'exit 0');
  command('stat', 'wc -c < "$3"');
  command('sha256sum', 'shasum -a 256 "$@"');
  command('systemctl', 'echo "$*" >> "$TEST_ROOT/calls"; case "$1" in stop) echo stopped > "$TEST_ROOT/state";; start) echo active > "$TEST_ROOT/state";; is-active) test "$(< "$TEST_ROOT/state")" = active;; esac');
  command('curl', 'if [[ $(< "$TEST_APP/bin/feetable") == bad-start ]]; then exit 22; fi; printf \'{"status":"ok"}/feetable/assets/app.js\\n\'');
  command('runuser', String.raw`shift 5
binary=$1; operation=$2; shift 2
if [[ $operation == backup ]]; then
  cp "$2" "$4"
elif [[ $operation == check ]]; then
  test "$(< "$binary")" != bad-check
else
  exit 64
fi`);
  let script = readFileSync(new URL('./deploy-release.sh', import.meta.url), 'utf8');
  script = script.replace('export PATH=/usr/sbin:/usr/bin:/sbin:/bin', 'export PATH="' + mocks + ':$PATH"')
    .replace('$EUID -ne 0 || ', '')
    .replace('app=/opt/feetable', 'app="' + app + '"')
    .replace('/run/lock/feetable-deploy.lock', join(root, 'lock'));
  const path = join(root, 'deploy.sh');
  writeFileSync(path, script);
  return { root, app, path, cleanup: () => rmSync(root, { recursive: true, force: true }) };
}

for (const scenario of ['success', 'bad-check', 'bad-start', 'bad-hash']) {
  test(scenario, () => {
    const f = fixture();
    try {
      const binary = scenario === 'success' ? 'new-program' : scenario;
      const hash = scenario === 'bad-hash' ? '0'.repeat(64) : createHash('sha256').update(binary).digest('hex');
      const result = spawnSync('bash', [f.path, 'a'.repeat(40), hash], {
        input: binary, encoding: 'utf8', env: { ...process.env, TEST_ROOT: f.root, TEST_APP: f.app },
      });
      assert.equal(result.status === 0, scenario === 'success', result.stderr);
      assert.equal(readFileSync(join(f.app, 'data/feetable.sqlite'), 'utf8'), 'synthetic-freight-data');
      assert.equal(readFileSync(join(f.root, 'state'), 'utf8').trim(), 'active');
      assert.equal(readFileSync(join(f.app, 'bin/feetable'), 'utf8'), scenario === 'success' ? 'new-program' : 'old-program');
      if (scenario === 'success') assert.equal(readFileSync(join(f.app, 'current-commit'), 'utf8').trim(), 'a'.repeat(40));
      if (scenario === 'bad-hash') assert.equal(existsSync(join(f.root, 'calls')), false);
      else {
        const backups = readdirSync(join(f.app, 'backups'));
        assert.equal(backups.length, 1);
        assert.equal(readFileSync(join(f.app, 'backups', backups[0]), 'utf8'), 'synthetic-freight-data');
      }
    } finally { f.cleanup(); }
  });
}
