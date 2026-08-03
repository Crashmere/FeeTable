package com.example.feetable.data.dao

import androidx.room.*
import com.example.feetable.data.entity.FeeTable
import kotlinx.coroutines.flow.Flow

@Dao
interface FeeTableDao {
    @Query("SELECT * FROM fee_tables ORDER BY updatedAt DESC")
    fun getAllTables(): Flow<List<FeeTable>>

    @Query("SELECT * FROM fee_tables WHERE id = :id")
    suspend fun getTableById(id: Int): FeeTable?

    @Insert
    suspend fun insert(table: FeeTable): Long

    @Update
    suspend fun update(table: FeeTable)

    @Delete
    suspend fun delete(table: FeeTable)

    @Query("UPDATE fee_tables SET updatedAt = :time WHERE id = :id")
    suspend fun updateTimestamp(id: Int, time: Long = System.currentTimeMillis())
}
