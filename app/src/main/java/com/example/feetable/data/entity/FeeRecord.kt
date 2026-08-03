package com.example.feetable.data.entity

import androidx.room.Entity
import androidx.room.ForeignKey
import androidx.room.Index
import androidx.room.PrimaryKey

@Entity(
    tableName = "fee_records",
    foreignKeys = [
        ForeignKey(
            entity = FeeTable::class,
            parentColumns = ["id"],
            childColumns = ["tableId"],
            onDelete = ForeignKey.CASCADE
        )
    ],
    indices = [Index("tableId")]
)
data class FeeRecord(
    @PrimaryKey(autoGenerate = true) val id: Int = 0,
    val tableId: Int,
    val day: Int,
    val location1: String,
    val location2: String,
    val quantity: Float,
    val unitPrice: Int? = null,
    val amount: Float,
    val tag: String? = null
)
