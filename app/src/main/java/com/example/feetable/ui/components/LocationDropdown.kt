package com.example.feetable.ui.components

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import com.example.feetable.data.entity.Location

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LocationDropdown(
    value: String,
    onValueChange: (String) -> Unit,
    locations: List<Location>,
    onAddLocation: (String) -> Unit,
    label: String,
    modifier: Modifier = Modifier
) {
    var expanded by remember { mutableStateOf(false) }
    var isFocused by remember { mutableStateOf(false) }
    var textFieldValue by remember(value) { mutableStateOf(value) }
    var showAddDialog by remember { mutableStateOf(false) }

    val filteredLocations = remember(textFieldValue, locations) {
        if (textFieldValue.isBlank()) locations
        else locations.filter { it.name.contains(textFieldValue, ignoreCase = true) }
    }

    ExposedDropdownMenuBox(
        expanded = expanded && isFocused,
        onExpandedChange = { expanded = it },
        modifier = modifier
    ) {
        OutlinedTextField(
            value = textFieldValue,
            onValueChange = {
                textFieldValue = it
                expanded = true
            },
            label = { Text(label) },
            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = expanded && isFocused) },
            singleLine = true,
            modifier = Modifier
                .menuAnchor()
                .fillMaxWidth()
                .onFocusChanged { focusState ->
                    isFocused = focusState.isFocused
                    if (!focusState.isFocused) {
                        expanded = false
                        if (textFieldValue != value) {
                            onValueChange(textFieldValue)
                        }
                    }
                }
        )

        val showAddOption = textFieldValue.isNotBlank() && filteredLocations.none { it.name == textFieldValue }

        ExposedDropdownMenu(
            expanded = expanded && isFocused && (filteredLocations.isNotEmpty() || showAddOption),
            onDismissRequest = { expanded = false },
            modifier = Modifier.heightIn(max = 200.dp)
        ) {
            filteredLocations.take(6).forEach { location ->
                DropdownMenuItem(
                    text = { Text(location.name) },
                    onClick = {
                        textFieldValue = location.name
                        onValueChange(location.name)
                        expanded = false
                    }
                )
            }

            if (textFieldValue.isNotBlank() && filteredLocations.none { it.name == textFieldValue }) {
                @Suppress("DEPRECATION")
                Divider()
                DropdownMenuItem(
                    text = {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Icon(
                                Icons.Default.Add,
                                contentDescription = null,
                                modifier = Modifier.size(18.dp)
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Text("添加「$textFieldValue」")
                        }
                    },
                    onClick = {
                        onAddLocation(textFieldValue)
                        onValueChange(textFieldValue)
                        expanded = false
                    }
                )
            }
        }
    }

    if (showAddDialog) {
        AddLocationDialog(
            onDismiss = { showAddDialog = false },
            onAdd = { name ->
                onAddLocation(name)
                textFieldValue = name
                onValueChange(name)
                showAddDialog = false
            }
        )
    }
}

@Composable
private fun AddLocationDialog(
    onDismiss: () -> Unit,
    onAdd: (String) -> Unit
) {
    var name by remember { mutableStateOf("") }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("添加新地点") },
        text = {
            OutlinedTextField(
                value = name,
                onValueChange = { name = it },
                label = { Text("地点名称") },
                singleLine = true,
                keyboardOptions = KeyboardOptions(imeAction = ImeAction.Done),
                keyboardActions = KeyboardActions(onDone = {
                    if (name.isNotBlank()) onAdd(name.trim())
                }),
                modifier = Modifier.fillMaxWidth()
            )
        },
        confirmButton = {
            TextButton(
                onClick = { if (name.isNotBlank()) onAdd(name.trim()) },
                enabled = name.isNotBlank()
            ) {
                Text("添加")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("取消")
            }
        }
    )
}
