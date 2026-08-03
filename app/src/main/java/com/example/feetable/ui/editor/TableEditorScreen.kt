package com.example.feetable.ui.editor

import androidx.compose.animation.animateContentSize
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.feetable.data.entity.FeeRecord
import com.example.feetable.data.entity.Location
import com.example.feetable.data.entity.Tag
import com.example.feetable.ui.components.LocationDropdown
import com.example.feetable.ui.components.TagDropdown

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TableEditorScreen(
    tableId: Int,
    onBack: () -> Unit,
    onExport: () -> Unit,
    onExportByTag: (String) -> Unit,
    viewModel: EditorViewModel = viewModel()
) {
    LaunchedEffect(tableId) {
        viewModel.loadTable(tableId)
    }

    val table by viewModel.table.collectAsState()
    val records by viewModel.records.collectAsState()
    val locations by viewModel.locations.collectAsState()
    val tags by viewModel.tags.collectAsState()

    var showAddSheet by remember { mutableStateOf(false) }
    var editingRecord by remember { mutableStateOf<FeeRecord?>(null) }
    var showEditYearMonth by remember { mutableStateOf(false) }
    var showTagExportDialog by remember { mutableStateOf(false) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    table?.let {
                        TextButton(onClick = { showEditYearMonth = true }) {
                            Text(
                                "${it.year}年${it.month}月",
                                style = MaterialTheme.typography.titleMedium
                            )
                            Icon(
                                Icons.Default.Edit,
                                contentDescription = "编辑",
                                modifier = Modifier.size(16.dp).padding(start = 4.dp)
                            )
                        }
                    }
                },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "返回")
                    }
                },
                actions = {
                    @Suppress("DEPRECATION")
                    IconButton(onClick = { showTagExportDialog = true }) {
                        Icon(Icons.Default.Label, contentDescription = "按标签汇总")
                    }
                    IconButton(onClick = onExport) {
                        Icon(Icons.Default.FileOpen, contentDescription = "导出")
                    }
                }
            )
        },
        floatingActionButton = {
            FloatingActionButton(onClick = { showAddSheet = true }) {
                Icon(Icons.Default.Add, contentDescription = "添加记录")
            }
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            if (records.isEmpty()) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        "暂无记录，点击右下角按钮添加",
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            } else {
                // Summary bar
                Surface(
                    modifier = Modifier.fillMaxWidth(),
                    tonalElevation = 1.dp
                ) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp, vertical = 8.dp),
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Text(
                            "共 ${records.size} 条记录",
                            style = MaterialTheme.typography.bodyMedium
                        )
                        Text(
                            "合计: ¥${String.format("%.2f", viewModel.getTotalAmount())}",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.primary
                        )
                    }
                }

                LazyColumn(
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    items(records, key = { it.id }) { record ->
                        RecordCard(
                            record = record,
                            month = table?.month ?: 1,
                            onEdit = { editingRecord = record },
                            onDelete = { viewModel.deleteRecord(record) }
                        )
                    }
                }
            }
        }

        if (showAddSheet) {
            RecordFormSheet(
                title = "添加记录",
                locations = locations,
                tags = tags,
                defaultDay = null,
                onDismiss = { showAddSheet = false },
                onSave = { day, loc1, loc2, qty, price, amount, tag ->
                    viewModel.addRecord(day, loc1, loc2, qty, price, amount, tag)
                    showAddSheet = false
                },
                onAddLocation = { viewModel.addLocation(it) },
                onAddTag = { viewModel.addTag(it) }
            )
        }

        editingRecord?.let { record ->
            RecordFormSheet(
                title = "编辑记录",
                locations = locations,
                tags = tags,
                initialRecord = record,
                defaultDay = record.day,
                onDismiss = { editingRecord = null },
                onSave = { day, loc1, loc2, qty, price, amount, tag ->
                    viewModel.updateRecord(
                        record.copy(
                            day = day,
                            location1 = loc1,
                            location2 = loc2,
                            quantity = qty,
                            unitPrice = price,
                            amount = amount,
                            tag = tag
                        )
                    )
                    editingRecord = null
                },
                onAddLocation = { viewModel.addLocation(it) },
                onAddTag = { viewModel.addTag(it) }
            )
        }

        if (showEditYearMonth) {
            table?.let { currentTable ->
                EditYearMonthDialog(
                    currentYear = currentTable.year,
                    currentMonth = currentTable.month,
                    onDismiss = { showEditYearMonth = false },
                    onSave = { year, month ->
                        viewModel.updateTable(year, month)
                        showEditYearMonth = false
                    }
                )
            }
        }

        if (showTagExportDialog) {
            val distinctTags = viewModel.getDistinctTags()
            if (distinctTags.isEmpty()) {
                AlertDialog(
                    onDismissRequest = { showTagExportDialog = false },
                    title = { Text("按标签汇总导出") },
                    text = { Text("当前表内没有带标签的记录，请先为记录添加标签。") },
                    confirmButton = {
                        TextButton(onClick = { showTagExportDialog = false }) {
                            Text("确定")
                        }
                    }
                )
            } else {
                TagExportDialog(
                    distinctTags = distinctTags,
                    onDismiss = { showTagExportDialog = false },
                    onSelectTag = { tag ->
                        showTagExportDialog = false
                        onExportByTag(tag)
                    }
                )
            }
        }
    }
}

@Composable
private fun TagExportDialog(
    distinctTags: List<String>,
    onDismiss: () -> Unit,
    onSelectTag: (String) -> Unit
) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("按标签汇总导出") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                Text(
                    "选择要汇总的标签：",
                    style = MaterialTheme.typography.bodyMedium,
                    modifier = Modifier.padding(bottom = 8.dp)
                )
                distinctTags.forEach { tag ->
                    TextButton(
                        onClick = { onSelectTag(tag) },
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Text(tag)
                    }
                }
            }
        },
        confirmButton = {},
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("取消")
            }
        }
    )
}

@Composable
private fun RecordCard(
    record: FeeRecord,
    month: Int,
    onEdit: () -> Unit,
    onDelete: () -> Unit
) {
    var showDeleteConfirm by remember { mutableStateOf(false) }

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .animateContentSize(),
        elevation = CardDefaults.cardElevation(defaultElevation = 1.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Surface(
                        shape = MaterialTheme.shapes.small,
                        tonalElevation = 4.dp
                    ) {
                        Text(
                            text = "${month}/${record.day}",
                            modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp),
                            style = MaterialTheme.typography.labelMedium
                        )
                    }
                    Text(
                        text = "${record.location1} → ${record.location2}",
                        style = MaterialTheme.typography.bodyMedium
                    )
                }
                Spacer(modifier = Modifier.height(4.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                    Text(
                        text = "数量: ${record.quantity}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    if (record.unitPrice != null) {
                        Text(
                            text = "单价: ${record.unitPrice}",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    } else {
                        Text(
                            text = "固定费用",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.tertiary
                        )
                    }
                    Text(
                        text = "¥${String.format("%.2f", record.amount)}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
                record.tag?.let { tagName ->
                    Spacer(modifier = Modifier.height(2.dp))
                    Surface(
                        shape = MaterialTheme.shapes.extraSmall,
                        color = MaterialTheme.colorScheme.secondaryContainer
                    ) {
                        Text(
                            text = tagName,
                            modifier = Modifier.padding(horizontal = 6.dp, vertical = 1.dp),
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onSecondaryContainer
                        )
                    }
                }
            }
            Column {
                IconButton(onClick = onEdit, modifier = Modifier.size(32.dp)) {
                    Icon(Icons.Default.Edit, contentDescription = "编辑", modifier = Modifier.size(18.dp))
                }
                IconButton(
                    onClick = { showDeleteConfirm = true },
                    modifier = Modifier.size(32.dp)
                ) {
                    Icon(
                        Icons.Default.Delete,
                        contentDescription = "删除",
                        modifier = Modifier.size(18.dp),
                        tint = MaterialTheme.colorScheme.error
                    )
                }
            }
        }
    }

    if (showDeleteConfirm) {
        AlertDialog(
            onDismissRequest = { showDeleteConfirm = false },
            title = { Text("确认删除") },
            text = { Text("确定要删除这条记录吗？") },
            confirmButton = {
                TextButton(onClick = {
                    onDelete()
                    showDeleteConfirm = false
                }) {
                    Text("删除", color = MaterialTheme.colorScheme.error)
                }
            },
            dismissButton = {
                TextButton(onClick = { showDeleteConfirm = false }) {
                    Text("取消")
                }
            }
        )
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun RecordFormSheet(
    title: String,
    locations: List<Location>,
    tags: List<Tag>,
    initialRecord: FeeRecord? = null,
    defaultDay: Int?,
    onDismiss: () -> Unit,
    onSave: (Int, String, String, Float, Int?, Float, String?) -> Unit,
    onAddLocation: (String) -> Unit,
    onAddTag: (String) -> Unit
) {
    var dayText by remember { mutableStateOf(defaultDay?.toString() ?: "") }
    var location1 by remember { mutableStateOf(initialRecord?.location1 ?: "") }
    var location2 by remember { mutableStateOf(initialRecord?.location2 ?: "") }
    var quantityText by remember { mutableStateOf(initialRecord?.quantity?.toString() ?: "") }
    var unitPriceText by remember { mutableStateOf(initialRecord?.unitPrice?.toString() ?: "") }
    var amountText by remember { mutableStateOf(initialRecord?.amount?.toString() ?: "") }
    var tagValue by remember { mutableStateOf(initialRecord?.tag ?: "") }

    val isFixedFee = unitPriceText.isBlank()
    val calculatedAmount = remember(quantityText, unitPriceText) {
        if (unitPriceText.isNotBlank()) {
            val qty = quantityText.toFloatOrNull() ?: 0f
            val price = unitPriceText.toIntOrNull() ?: 0
            qty * price
        } else null
    }

    LaunchedEffect(calculatedAmount) {
        calculatedAmount?.let {
            amountText = String.format("%.2f", it)
        }
    }

    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 24.dp)
                .padding(bottom = 32.dp)
                .imePadding(),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Text(
                text = title,
                style = MaterialTheme.typography.titleLarge,
                modifier = Modifier.padding(bottom = 8.dp)
            )

            OutlinedTextField(
                value = dayText,
                onValueChange = {
                    val filtered = it.filter { c -> c.isDigit() }
                    if (filtered.isEmpty() || (filtered.toIntOrNull() ?: 0) <= 31) {
                        dayText = filtered
                    }
                },
                label = { Text("日") },
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                singleLine = true,
                modifier = Modifier.fillMaxWidth()
            )

            LocationDropdown(
                value = location1,
                onValueChange = { location1 = it },
                locations = locations,
                onAddLocation = onAddLocation,
                label = "地点1",
                modifier = Modifier.fillMaxWidth()
            )

            LocationDropdown(
                value = location2,
                onValueChange = { location2 = it },
                locations = locations,
                onAddLocation = onAddLocation,
                label = "地点2",
                modifier = Modifier.fillMaxWidth()
            )

            OutlinedTextField(
                value = quantityText,
                onValueChange = { quantityText = it },
                label = { Text("运输数量") },
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal),
                singleLine = true,
                modifier = Modifier.fillMaxWidth()
            )

            OutlinedTextField(
                value = unitPriceText,
                onValueChange = { unitPriceText = it.filter { c -> c.isDigit() } },
                label = { Text("单价 (留空表示固定费用)") },
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                singleLine = true,
                modifier = Modifier.fillMaxWidth()
            )

            OutlinedTextField(
                value = amountText,
                onValueChange = { if (isFixedFee) amountText = it },
                label = { Text(if (isFixedFee) "运输金额 (手动输入)" else "运输金额 (自动计算)") },
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal),
                singleLine = true,
                enabled = isFixedFee,
                modifier = Modifier.fillMaxWidth()
            )

            TagDropdown(
                value = tagValue,
                onValueChange = { tagValue = it },
                tags = tags,
                onAddTag = onAddTag,
                modifier = Modifier.fillMaxWidth()
            )

            Spacer(modifier = Modifier.height(8.dp))

            Button(
                onClick = {
                    val day = dayText.toIntOrNull() ?: return@Button
                    val qty = quantityText.toFloatOrNull() ?: return@Button
                    val price = unitPriceText.toIntOrNull()
                    val amount = amountText.toFloatOrNull() ?: return@Button
                    if (location1.isNotBlank() && location2.isNotBlank() && day in 1..31) {
                        onSave(day, location1.trim(), location2.trim(), qty, price, amount,
                            tagValue.takeIf { it.isNotBlank() })
                    }
                },
                modifier = Modifier.fillMaxWidth(),
                enabled = dayText.isNotBlank() && location1.isNotBlank() &&
                        location2.isNotBlank() && quantityText.isNotBlank() &&
                        amountText.isNotBlank()
            ) {
                Text("保存")
            }
        }
    }
}

@Composable
private fun EditYearMonthDialog(
    currentYear: Int,
    currentMonth: Int,
    onDismiss: () -> Unit,
    onSave: (Int, Int) -> Unit
) {
    var yearText by remember { mutableStateOf(currentYear.toString()) }
    var monthText by remember { mutableStateOf(currentMonth.toString()) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("修改年月") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                OutlinedTextField(
                    value = yearText,
                    onValueChange = { yearText = it.filter { c -> c.isDigit() } },
                    label = { Text("年份") },
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth()
                )
                OutlinedTextField(
                    value = monthText,
                    onValueChange = {
                        val filtered = it.filter { c -> c.isDigit() }
                        if (filtered.isEmpty() || (filtered.toIntOrNull() ?: 0) <= 12) {
                            monthText = filtered
                        }
                    },
                    label = { Text("月份 (1-12)") },
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth()
                )
            }
        },
        confirmButton = {
            TextButton(
                onClick = {
                    val year = yearText.toIntOrNull() ?: return@TextButton
                    val month = monthText.toIntOrNull() ?: return@TextButton
                    if (month in 1..12) onSave(year, month)
                }
            ) {
                Text("保存")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("取消")
            }
        }
    )
}
