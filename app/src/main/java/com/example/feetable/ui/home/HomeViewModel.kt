package com.example.feetable.ui.home

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.feetable.FeeTableApplication
import com.example.feetable.data.entity.FeeTable
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import java.util.Calendar

data class TableSummary(
    val table: FeeTable,
    val recordCount: Int,
    val totalAmount: Float
)

class HomeViewModel(application: Application) : AndroidViewModel(application) {
    private val repository = (application as FeeTableApplication).repository

    val tables: StateFlow<List<FeeTable>> = repository.allTables
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    fun createTable(year: Int, month: Int) {
        viewModelScope.launch {
            repository.insertTable(FeeTable(year = year, month = month))
        }
    }

    fun deleteTable(table: FeeTable) {
        viewModelScope.launch {
            repository.deleteTable(table)
        }
    }

    fun getRecordCount(tableId: Int): Flow<Int> = repository.getRecordCount(tableId)

    fun getTotalAmount(tableId: Int): Flow<Float?> = repository.getTotalAmount(tableId)

    fun getCurrentYear(): Int = Calendar.getInstance().get(Calendar.YEAR)
    fun getCurrentMonth(): Int = Calendar.getInstance().get(Calendar.MONTH) + 1
}
