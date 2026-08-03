package com.example.feetable.export

import android.content.Context
import android.graphics.Bitmap
import com.example.feetable.data.entity.FeeRecord
import com.example.feetable.data.entity.FeeTable
import java.io.File
import java.io.FileOutputStream

class PngExporter(
    private val context: Context,
    private val table: FeeTable,
    private val records: List<FeeRecord>
) {
    fun export(): File {
        val renderer = TableRenderer(table, records)
        val bitmap = renderer.render()

        val file = File(context.cacheDir, "运费明细表_${table.year}年${table.month}月.png")
        FileOutputStream(file).use { out ->
            bitmap.compress(Bitmap.CompressFormat.PNG, 100, out)
        }
        bitmap.recycle()

        return file
    }
}
