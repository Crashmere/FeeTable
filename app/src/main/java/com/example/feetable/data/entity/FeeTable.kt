package com.example.feetable.data.entity

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "fee_tables")
data class FeeTable(
    @PrimaryKey(autoGenerate = true) val id: Int = 0,
    val year: Int,
    val month: Int,
    val createdAt: Long = System.currentTimeMillis(),
    val updatedAt: Long = System.currentTimeMillis()
)
