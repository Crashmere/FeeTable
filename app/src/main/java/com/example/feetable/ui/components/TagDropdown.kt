package com.example.feetable.ui.components

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import com.example.feetable.data.entity.Tag

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TagDropdown(
    value: String,
    onValueChange: (String) -> Unit,
    tags: List<Tag>,
    onAddTag: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    var expanded by remember { mutableStateOf(false) }
    var isFocused by remember { mutableStateOf(false) }
    var textFieldValue by remember(value) { mutableStateOf(value) }

    val filteredTags = remember(textFieldValue, tags) {
        if (textFieldValue.isBlank()) tags
        else tags.filter { it.name.contains(textFieldValue, ignoreCase = true) }
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
            label = { Text("标签（可选）") },
            trailingIcon = {
                if (textFieldValue.isNotBlank()) {
                    IconButton(onClick = {
                        textFieldValue = ""
                        onValueChange("")
                    }) {
                        Icon(Icons.Default.Close, contentDescription = "清除")
                    }
                } else {
                    ExposedDropdownMenuDefaults.TrailingIcon(expanded = expanded && isFocused)
                }
            },
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

        val showAddOption = textFieldValue.isNotBlank() && filteredTags.none { it.name == textFieldValue }

        ExposedDropdownMenu(
            expanded = expanded && isFocused && (filteredTags.isNotEmpty() || showAddOption),
            onDismissRequest = { expanded = false },
            modifier = Modifier.heightIn(max = 200.dp)
        ) {
            filteredTags.take(6).forEach { tag ->
                DropdownMenuItem(
                    text = { Text(tag.name) },
                    onClick = {
                        textFieldValue = tag.name
                        onValueChange(tag.name)
                        expanded = false
                    }
                )
            }

            if (textFieldValue.isNotBlank() && filteredTags.none { it.name == textFieldValue }) {
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
                        onAddTag(textFieldValue)
                        onValueChange(textFieldValue)
                        expanded = false
                    }
                )
            }
        }
    }
}
