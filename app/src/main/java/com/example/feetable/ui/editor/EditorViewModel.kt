package com.example.feetable.ui.editor

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.feetable.FeeTableApplication
import com.example.feetable.data.entity.FeeRecord
import com.example.feetable.data.entity.FeeTable
import com.example.feetable.data.entity.Location
import com.example.feetable.data.entity.Tag
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch

class EditorViewModel(application: Application) : AndroidViewModel(application) {
    private val repository = (application as FeeTableApplication).repository

    private val _table = MutableStateFlow<FeeTable?>(null)
    val table: StateFlow<FeeTable?> = _table.asStateFlow()

    private val _records = MutableStateFlow<List<FeeRecord>>(emptyList())
    val records: StateFlow<List<FeeRecord>> = _records.asStateFlow()

    val locations: StateFlow<List<Location>> = repository.allLocations
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    val tags: StateFlow<List<Tag>> = repository.allTags
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    private var _tableId: Int = 0

    fun loadTable(tableId: Int) {
        _tableId = tableId
        viewModelScope.launch {
            _table.value = repository.getTableById(tableId)
        }
        viewModelScope.launch {
            repository.getRecordsByTableId(tableId).collect {
                _records.value = it
            }
        }
    }

    fun updateTable(year: Int, month: Int) {
        viewModelScope.launch {
            _table.value?.let { currentTable ->
                val updated = currentTable.copy(
                    year = year,
                    month = month,
                    updatedAt = System.currentTimeMillis()
                )
                repository.updateTable(updated)
                _table.value = updated
            }
        }
    }

    fun addRecord(
        day: Int,
        location1: String,
        location2: String,
        quantity: Float,
        unitPrice: Int?,
        amount: Float,
        tag: String? = null
    ) {
        viewModelScope.launch {
            val record = FeeRecord(
                tableId = _tableId,
                day = day,
                location1 = location1,
                location2 = location2,
                quantity = quantity,
                unitPrice = unitPrice,
                amount = amount,
                tag = tag?.takeIf { it.isNotBlank() }
            )
            repository.insertRecord(record)
        }
    }

    fun updateRecord(record: FeeRecord) {
        viewModelScope.launch {
            repository.updateRecord(record)
        }
    }

    fun deleteRecord(record: FeeRecord) {
        viewModelScope.launch {
            repository.deleteRecord(record)
        }
    }

    fun addLocation(name: String) {
        viewModelScope.launch {
            repository.insertLocation(name)
        }
    }

    fun addTag(name: String) {
        viewModelScope.launch {
            repository.insertTag(name)
        }
    }

    fun getDistinctTags(): List<String> {
        return _records.value.mapNotNull { it.tag }.distinct()
    }

    fun getTotalAmount(): Float {
        return _records.value.sumOf { it.amount.toDouble() }.toFloat()
    }
}
