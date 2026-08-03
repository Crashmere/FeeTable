package com.example.feetable.export

import android.content.Context
import com.example.feetable.data.entity.FeeRecord
import com.example.feetable.data.entity.FeeTable
import org.apache.poi.ss.usermodel.*
import org.apache.poi.ss.util.CellRangeAddress
import org.apache.poi.xssf.usermodel.XSSFWorkbook
import java.io.File
import java.io.FileOutputStream

class ExcelExporter(
    private val context: Context,
    private val table: FeeTable,
    private val records: List<FeeRecord>
) {
    fun export(): File {
        val workbook = XSSFWorkbook()
        val sheet = workbook.createSheet("运费明细表")

        val borderStyle = workbook.createCellStyle().apply {
            borderTop = BorderStyle.THIN
            borderBottom = BorderStyle.THIN
            borderLeft = BorderStyle.THIN
            borderRight = BorderStyle.THIN
            setAlignment(HorizontalAlignment.CENTER)
            setVerticalAlignment(VerticalAlignment.CENTER)
        }

        val titleStyle = workbook.createCellStyle().apply {
            cloneStyleFrom(borderStyle)
            val font = workbook.createFont()
            font.bold = true
            font.fontHeightInPoints = 14
            setFont(font)
        }

        val headerStyle = workbook.createCellStyle().apply {
            cloneStyleFrom(borderStyle)
            val font = workbook.createFont()
            font.bold = true
            setFont(font)
        }

        val numberStyle = workbook.createCellStyle().apply {
            cloneStyleFrom(borderStyle)
            setAlignment(HorizontalAlignment.RIGHT)
            dataFormat = workbook.createDataFormat().getFormat("0.000")
        }

        val amountStyle = workbook.createCellStyle().apply {
            cloneStyleFrom(borderStyle)
            setAlignment(HorizontalAlignment.RIGHT)
            dataFormat = workbook.createDataFormat().getFormat("0.00")
        }

        val intStyle = workbook.createCellStyle().apply {
            cloneStyleFrom(borderStyle)
            setAlignment(HorizontalAlignment.CENTER)
        }

        var rowIdx = 0

        // Title row
        val titleRow = sheet.createRow(rowIdx)
        titleRow.createCell(0).apply {
            setCellValue("运费明细表")
            cellStyle = titleStyle
        }
        sheet.addMergedRegion(CellRangeAddress(rowIdx, rowIdx, 0, 6))
        for (i in 1..6) titleRow.createCell(i).cellStyle = titleStyle
        rowIdx++

        // Header row 1
        val headerRow1 = sheet.createRow(rowIdx)
        headerRow1.createCell(0).apply {
            setCellValue("${table.year}年")
            cellStyle = headerStyle
        }
        headerRow1.createCell(1).cellStyle = headerStyle
        sheet.addMergedRegion(CellRangeAddress(rowIdx, rowIdx, 0, 1))

        headerRow1.createCell(2).apply {
            setCellValue("摘要")
            cellStyle = headerStyle
        }
        headerRow1.createCell(3).cellStyle = headerStyle
        sheet.addMergedRegion(CellRangeAddress(rowIdx, rowIdx, 2, 3))

        headerRow1.createCell(4).apply {
            setCellValue("运输数量")
            cellStyle = headerStyle
        }
        sheet.addMergedRegion(CellRangeAddress(rowIdx, rowIdx + 1, 4, 4))

        headerRow1.createCell(5).apply {
            setCellValue("单价")
            cellStyle = headerStyle
        }
        sheet.addMergedRegion(CellRangeAddress(rowIdx, rowIdx + 1, 5, 5))

        headerRow1.createCell(6).apply {
            setCellValue("运输金额")
            cellStyle = headerStyle
        }
        sheet.addMergedRegion(CellRangeAddress(rowIdx, rowIdx + 1, 6, 6))
        rowIdx++

        // Header row 2
        val headerRow2 = sheet.createRow(rowIdx)
        headerRow2.createCell(0).apply {
            setCellValue("月")
            cellStyle = headerStyle
        }
        headerRow2.createCell(1).apply {
            setCellValue("日")
            cellStyle = headerStyle
        }
        headerRow2.createCell(2).cellStyle = headerStyle
        headerRow2.createCell(3).cellStyle = headerStyle
        headerRow2.createCell(4).cellStyle = headerStyle
        headerRow2.createCell(5).cellStyle = headerStyle
        headerRow2.createCell(6).cellStyle = headerStyle
        rowIdx++

        // Data rows
        for (record in records) {
            val row = sheet.createRow(rowIdx)
            row.createCell(0).apply {
                setCellValue(table.month.toDouble())
                cellStyle = intStyle
            }
            row.createCell(1).apply {
                setCellValue(record.day.toDouble())
                cellStyle = intStyle
            }
            row.createCell(2).apply {
                setCellValue(record.location1)
                cellStyle = borderStyle
            }
            row.createCell(3).apply {
                setCellValue(record.location2)
                cellStyle = borderStyle
            }
            row.createCell(4).apply {
                setCellValue(record.quantity.toDouble())
                cellStyle = numberStyle
            }
            row.createCell(5).apply {
                if (record.unitPrice != null) {
                    setCellValue(record.unitPrice.toDouble())
                }
                cellStyle = intStyle
            }
            row.createCell(6).apply {
                setCellValue(record.amount.toDouble())
                cellStyle = amountStyle
            }
            rowIdx++
        }

        // Total row
        val totalRow = sheet.createRow(rowIdx)
        for (i in 0..5) totalRow.createCell(i).cellStyle = borderStyle
        sheet.addMergedRegion(CellRangeAddress(rowIdx, rowIdx, 0, 5))
        totalRow.createCell(6).apply {
            val total = records.sumOf { it.amount.toDouble() }
            setCellValue(total)
            cellStyle = amountStyle
        }

        // Auto-size columns
        for (i in 0..6) {
            sheet.setColumnWidth(i, 4000)
        }
        sheet.setColumnWidth(2, 5000)
        sheet.setColumnWidth(3, 5000)
        sheet.setColumnWidth(6, 5000)

        val file = File(context.cacheDir, "运费明细表_${table.year}年${table.month}月.xlsx")
        FileOutputStream(file).use { out ->
            workbook.write(out)
        }
        workbook.close()

        return file
    }
}
