package com.example.feetable.ui.home

import androidx.compose.animation.animateContentSize
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.feetable.data.entity.FeeTable

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    onTableClick: (Int) -> Unit,
    viewModel: HomeViewModel = viewModel()
) {
    val tables by viewModel.tables.collectAsState()
    var showCreateDialog by remember { mutableStateOf(false) }
    var tableToDelete by remember { mutableStateOf<FeeTable?>(null) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("运费明细表") }
            )
        },
        floatingActionButton = {
            FloatingActionButton(onClick = { showCreateDialog = true }) {
                Icon(Icons.Default.Add, contentDescription = "新建表格")
            }
        }
    ) { padding ->
        if (tables.isEmpty()) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = "暂无表格，点击右下角按钮创建",
                    style = MaterialTheme.typography.bodyLarge,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                items(tables, key = { it.id }) { table ->
                    TableCard(
                        table = table,
                        viewModel = viewModel,
                        onClick = { onTableClick(table.id) },
                        onDelete = { tableToDelete = table }
                    )
                }
            }
        }

        if (showCreateDialog) {
            CreateTableDialog(
                defaultYear = viewModel.getCurrentYear(),
                defaultMonth = viewModel.getCurrentMonth(),
                onDismiss = { showCreateDialog = false },
                onCreate = { year, month ->
                    viewModel.createTable(year, month)
                    showCreateDialog = false
                }
            )
        }

        tableToDelete?.let { table ->
            AlertDialog(
                onDismissRequest = { tableToDelete = null },
                title = { Text("确认删除") },
                text = { Text("确定要删除 ${table.year}年${table.month}月 的表格吗？所有记录将被一并删除。") },
                confirmButton = {
                    TextButton(onClick = {
                        viewModel.deleteTable(table)
                        tableToDelete = null
                    }) {
                        Text("删除", color = MaterialTheme.colorScheme.error)
                    }
                },
                dismissButton = {
                    TextButton(onClick = { tableToDelete = null }) {
                        Text("取消")
                    }
                }
            )
        }
    }
}

@Composable
private fun TableCard(
    table: FeeTable,
    viewModel: HomeViewModel,
    onClick: () -> Unit,
    onDelete: () -> Unit
) {
    val recordCount by viewModel.getRecordCount(table.id).collectAsState(initial = 0)
    val totalAmount by viewModel.getTotalAmount(table.id).collectAsState(initial = null)

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .animateContentSize()
            .clickable(onClick = onClick),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = "${table.year}年${table.month}月",
                    style = MaterialTheme.typography.titleMedium
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "${recordCount}条记录  |  合计: ¥${String.format("%.2f", totalAmount ?: 0f)}",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
            IconButton(onClick = onDelete) {
                Icon(
                    Icons.Default.Delete,
                    contentDescription = "删除",
                    tint = MaterialTheme.colorScheme.error
                )
            }
        }
    }
}

@Composable
private fun CreateTableDialog(
    defaultYear: Int,
    defaultMonth: Int,
    onDismiss: () -> Unit,
    onCreate: (Int, Int) -> Unit
) {
    var yearText by remember { mutableStateOf(defaultYear.toString()) }
    var monthText by remember { mutableStateOf(defaultMonth.toString()) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("新建运费表") },
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
                    val year = yearText.toIntOrNull() ?: defaultYear
                    val month = monthText.toIntOrNull() ?: defaultMonth
                    if (month in 1..12) {
                        onCreate(year, month)
                    }
                }
            ) {
                Text("创建")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("取消")
            }
        }
    )
}
