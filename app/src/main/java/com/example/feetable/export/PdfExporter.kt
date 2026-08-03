package com.example.feetable.export

import android.content.Context
import android.graphics.pdf.PdfDocument
import com.example.feetable.data.entity.FeeRecord
import com.example.feetable.data.entity.FeeTable
import java.io.File
import java.io.FileOutputStream

class PdfExporter(
    private val context: Context,
    private val table: FeeTable,
    private val records: List<FeeRecord>
) {
    fun export(): File {
        val renderer = TableRenderer(table, records)
        val width = renderer.getWidth().toInt()
        val height = renderer.getHeight().toInt()

        val document = PdfDocument()
        val pageInfo = PdfDocument.PageInfo.Builder(width, height, 1).create()
        val page = document.startPage(pageInfo)

        renderer.drawOnCanvas(page.canvas)

        document.finishPage(page)

        val file = File(context.cacheDir, "运费明细表_${table.year}年${table.month}月.pdf")
        FileOutputStream(file).use { out ->
            document.writeTo(out)
        }
        document.close()

        return file
    }
}
