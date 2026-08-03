package com.example.feetable.data.dao

import androidx.room.*
import com.example.feetable.data.entity.FeeRecord
import kotlinx.coroutines.flow.Flow

@Dao
interface FeeRecordDao {
    @Query("SELECT * FROM fee_records WHERE tableId = :tableId ORDER BY day ASC, id ASC")
    fun getRecordsByTableId(tableId: Int): Flow<List<FeeRecord>>

    @Query("SELECT * FROM fee_records WHERE tableId = :tableId ORDER BY day ASC, id ASC")
    suspend fun getRecordsByTableIdOnce(tableId: Int): List<FeeRecord>

    @Query("SELECT SUM(amount) FROM fee_records WHERE tableId = :tableId")
    fun getTotalAmount(tableId: Int): Flow<Float?>

    @Insert
    suspend fun insert(record: FeeRecord): Long

    @Update
    suspend fun update(record: FeeRecord)

    @Delete
    suspend fun delete(record: FeeRecord)

    @Query("SELECT COUNT(*) FROM fee_records WHERE tableId = :tableId")
    fun getRecordCount(tableId: Int): Flow<Int>
}
