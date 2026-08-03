package com.example.feetable.data.dao

import androidx.room.*
import com.example.feetable.data.entity.Tag
import kotlinx.coroutines.flow.Flow

@Dao
interface TagDao {
    @Query("SELECT * FROM tags ORDER BY selectionCount DESC, name ASC")
    fun getAllOrderByCount(): Flow<List<Tag>>

    @Insert(onConflict = OnConflictStrategy.IGNORE)
    suspend fun insert(tag: Tag): Long

    @Query("UPDATE tags SET selectionCount = selectionCount + 1 WHERE name = :name")
    suspend fun incrementCount(name: String)

    @Delete
    suspend fun delete(tag: Tag)
}
