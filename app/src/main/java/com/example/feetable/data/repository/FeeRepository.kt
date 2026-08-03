package com.example.feetable.data.repository

import com.example.feetable.data.dao.FeeRecordDao
import com.example.feetable.data.dao.FeeTableDao
import com.example.feetable.data.dao.LocationDao
import com.example.feetable.data.dao.TagDao
import com.example.feetable.data.entity.FeeRecord
import com.example.feetable.data.entity.FeeTable
import com.example.feetable.data.entity.Location
import com.example.feetable.data.entity.Tag
import kotlinx.coroutines.flow.Flow

class FeeRepository(
    private val feeTableDao: FeeTableDao,
    private val feeRecordDao: FeeRecordDao,
    private val locationDao: LocationDao,
    private val tagDao: TagDao
) {
    // FeeTable operations
    val allTables: Flow<List<FeeTable>> = feeTableDao.getAllTables()

    suspend fun getTableById(id: Int): FeeTable? = feeTableDao.getTableById(id)

    suspend fun insertTable(table: FeeTable): Long = feeTableDao.insert(table)

    suspend fun updateTable(table: FeeTable) = feeTableDao.update(table)

    suspend fun deleteTable(table: FeeTable) = feeTableDao.delete(table)

    // FeeRecord operations
    fun getRecordsByTableId(tableId: Int): Flow<List<FeeRecord>> =
        feeRecordDao.getRecordsByTableId(tableId)

    suspend fun getRecordsByTableIdOnce(tableId: Int): List<FeeRecord> =
        feeRecordDao.getRecordsByTableIdOnce(tableId)

    fun getTotalAmount(tableId: Int): Flow<Float?> = feeRecordDao.getTotalAmount(tableId)

    fun getRecordCount(tableId: Int): Flow<Int> = feeRecordDao.getRecordCount(tableId)

    suspend fun insertRecord(record: FeeRecord): Long {
        val id = feeRecordDao.insert(record)
        feeTableDao.updateTimestamp(record.tableId)
        locationDao.insert(Location(name = record.location1))
        locationDao.incrementCount(record.location1)
        locationDao.insert(Location(name = record.location2))
        locationDao.incrementCount(record.location2)
        record.tag?.let {
            tagDao.insert(Tag(name = it))
            tagDao.incrementCount(it)
        }
        return id
    }

    suspend fun updateRecord(record: FeeRecord) {
        feeRecordDao.update(record)
        feeTableDao.updateTimestamp(record.tableId)
    }

    suspend fun deleteRecord(record: FeeRecord) {
        feeRecordDao.delete(record)
        feeTableDao.updateTimestamp(record.tableId)
    }

    // Location operations
    val allLocations: Flow<List<Location>> = locationDao.getAllOrderByCount()

    fun searchLocations(query: String): Flow<List<Location>> = locationDao.search(query)

    suspend fun insertLocation(name: String): Long =
        locationDao.insert(Location(name = name))

    suspend fun deleteLocation(location: Location) = locationDao.delete(location)

    // Tag operations
    val allTags: Flow<List<Tag>> = tagDao.getAllOrderByCount()

    suspend fun insertTag(name: String): Long =
        tagDao.insert(Tag(name = name))

    suspend fun deleteTag(tag: Tag) = tagDao.delete(tag)
}
