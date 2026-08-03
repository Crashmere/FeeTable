package com.example.feetable.data.dao

import androidx.room.*
import com.example.feetable.data.entity.Location
import kotlinx.coroutines.flow.Flow

@Dao
interface LocationDao {
    @Query("SELECT * FROM locations ORDER BY selectionCount DESC, name ASC")
    fun getAllOrderByCount(): Flow<List<Location>>

    @Query("SELECT * FROM locations WHERE name LIKE '%' || :query || '%' ORDER BY selectionCount DESC, name ASC")
    fun search(query: String): Flow<List<Location>>

    @Insert(onConflict = OnConflictStrategy.IGNORE)
    suspend fun insert(location: Location): Long

    @Query("UPDATE locations SET selectionCount = selectionCount + 1 WHERE name = :name")
    suspend fun incrementCount(name: String)

    @Delete
    suspend fun delete(location: Location)
}
