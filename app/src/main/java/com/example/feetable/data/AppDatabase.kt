package com.example.feetable.data

import android.content.Context
import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import androidx.room.migration.Migration
import androidx.sqlite.db.SupportSQLiteDatabase
import com.example.feetable.data.dao.FeeRecordDao
import com.example.feetable.data.dao.FeeTableDao
import com.example.feetable.data.dao.LocationDao
import com.example.feetable.data.dao.TagDao
import com.example.feetable.data.entity.FeeRecord
import com.example.feetable.data.entity.FeeTable
import com.example.feetable.data.entity.Location
import com.example.feetable.data.entity.Tag

@Database(
    entities = [FeeTable::class, FeeRecord::class, Location::class, Tag::class],
    version = 2,
    exportSchema = false
)
abstract class AppDatabase : RoomDatabase() {
    abstract fun feeTableDao(): FeeTableDao
    abstract fun feeRecordDao(): FeeRecordDao
    abstract fun locationDao(): LocationDao
    abstract fun tagDao(): TagDao

    companion object {
        @Volatile
        private var INSTANCE: AppDatabase? = null

        private val MIGRATION_1_2 = object : Migration(1, 2) {
            override fun migrate(db: SupportSQLiteDatabase) {
                db.execSQL("ALTER TABLE fee_records ADD COLUMN tag TEXT DEFAULT NULL")
                db.execSQL("CREATE TABLE IF NOT EXISTS tags (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, name TEXT NOT NULL, selectionCount INTEGER NOT NULL DEFAULT 0)")
                db.execSQL("CREATE UNIQUE INDEX IF NOT EXISTS index_tags_name ON tags (name)")
            }
        }

        fun getDatabase(context: Context): AppDatabase {
            return INSTANCE ?: synchronized(this) {
                val instance = Room.databaseBuilder(
                    context.applicationContext,
                    AppDatabase::class.java,
                    "fee_table_database"
                )
                    .addMigrations(MIGRATION_1_2)
                    .build()
                INSTANCE = instance
                instance
            }
        }
    }
}
