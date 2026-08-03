package com.example.feetable.export

import android.graphics.*
import com.example.feetable.data.entity.FeeRecord
import com.example.feetable.data.entity.FeeTable

class TableRenderer(
    private val table: FeeTable,
    private val records: List<FeeRecord>
) {
    companion object {
        private const val PADDING = 40f
        private const val CELL_PADDING = 12f
        private const val TITLE_FONT_SIZE = 28f
        private const val HEADER_FONT_SIZE = 20f
        private const val BODY_FONT_SIZE = 18f
        private const val ROW_HEIGHT = 48f
        private const val TITLE_ROW_HEIGHT = 56f
    }

    private val borderPaint = Paint().apply {
        style = Paint.Style.STROKE
        strokeWidth = 1.5f
        color = Color.BLACK
        isAntiAlias = true
    }

    private val titlePaint = Paint().apply {
        textSize = TITLE_FONT_SIZE
        color = Color.BLACK
        isAntiAlias = true
        typeface = Typeface.DEFAULT_BOLD
        textAlign = Paint.Align.CENTER
    }

    private val headerPaint = Paint().apply {
        textSize = HEADER_FONT_SIZE
        color = Color.BLACK
        isAntiAlias = true
        typeface = Typeface.DEFAULT_BOLD
        textAlign = Paint.Align.CENTER
    }

    private val bodyPaint = Paint().apply {
        textSize = BODY_FONT_SIZE
        color = Color.BLACK
        isAntiAlias = true
    }

    private val bodyRightPaint = Paint().apply {
        textSize = BODY_FONT_SIZE
        color = Color.BLACK
        isAntiAlias = true
        textAlign = Paint.Align.RIGHT
    }

    private val bodyCenterPaint = Paint().apply {
        textSize = BODY_FONT_SIZE
        color = Color.BLACK
        isAntiAlias = true
        textAlign = Paint.Align.CENTER
    }

    private val colWidths = calculateColumnWidths()

    private fun calculateColumnWidths(): FloatArray {
        val minWidths = floatArrayOf(60f, 60f, 120f, 120f, 120f, 80f, 140f)

        // Check actual content widths
        val measurePaint = Paint().apply { textSize = BODY_FONT_SIZE }

        for (record in records) {
            val texts = arrayOf(
                table.month.toString(),
                record.day.toString(),
                record.location1,
                record.location2,
                String.format("%.3f", record.quantity),
                record.unitPrice?.toString() ?: "",
                String.format("%.2f", record.amount)
            )
            for (i in texts.indices) {
                val w = measurePaint.measureText(texts[i]) + CELL_PADDING * 2
                if (w > minWidths[i]) minWidths[i] = w
            }
        }

        // Also check header text widths
        val headers = arrayOf("月", "日", "", "", "运输数量", "单价", "运输金额")
        for (i in headers.indices) {
            if (headers[i].isNotEmpty()) {
                val w = headerPaint.measureText(headers[i]) + CELL_PADDING * 2
                if (w > minWidths[i]) minWidths[i] = w
            }
        }

        return minWidths
    }

    private fun getTotalWidth(): Float = colWidths.sum() + PADDING * 2

    private fun getTotalHeight(): Float {
        val dataRows = records.size + 1 // +1 for total row
        return PADDING * 2 + TITLE_ROW_HEIGHT + ROW_HEIGHT * 2 + ROW_HEIGHT * dataRows
    }

    fun render(): Bitmap {
        val width = getTotalWidth().toInt()
        val height = getTotalHeight().toInt()
        val bitmap = Bitmap.createBitmap(width, height, Bitmap.Config.ARGB_8888)
        val canvas = Canvas(bitmap)

        canvas.drawColor(Color.WHITE)
        drawTable(canvas)

        return bitmap
    }

    fun drawOnCanvas(canvas: Canvas, offsetX: Float = 0f, offsetY: Float = 0f) {
        canvas.save()
        canvas.translate(offsetX, offsetY)
        drawTable(canvas)
        canvas.restore()
    }

    fun getWidth(): Float = getTotalWidth()
    fun getHeight(): Float = getTotalHeight()

    private fun drawTable(canvas: Canvas) {
        val startX = PADDING
        val startY = PADDING
        val tableWidth = colWidths.sum()

        var y = startY

        // Title row: "运费明细表"
        canvas.drawRect(startX, y, startX + tableWidth, y + TITLE_ROW_HEIGHT, borderPaint)
        canvas.drawText(
            "运费明细表",
            startX + tableWidth / 2,
            y + TITLE_ROW_HEIGHT / 2 + titlePaint.textSize / 3,
            titlePaint
        )
        y += TITLE_ROW_HEIGHT

        // Header row 1: "YYYY年" (spans col 0-1), "摘要" (spans col 2-3), "运输数量", "单价", "运输金额"
        var x = startX

        // "YYYY年" spanning columns 0 and 1
        val yearWidth = colWidths[0] + colWidths[1]
        canvas.drawRect(x, y, x + yearWidth, y + ROW_HEIGHT, borderPaint)
        canvas.drawText(
            "${table.year}年",
            x + yearWidth / 2,
            y + ROW_HEIGHT / 2 + headerPaint.textSize / 3,
            headerPaint
        )
        x += yearWidth

        // "摘要" spanning columns 2 and 3
        val summaryWidth = colWidths[2] + colWidths[3]
        canvas.drawRect(x, y, x + summaryWidth, y + ROW_HEIGHT, borderPaint)
        canvas.drawText(
            "摘要",
            x + summaryWidth / 2,
            y + ROW_HEIGHT / 2 + headerPaint.textSize / 3,
            headerPaint
        )
        x += summaryWidth

        // "运输数量"
        canvas.drawRect(x, y, x + colWidths[4], y + ROW_HEIGHT * 2, borderPaint)
        canvas.drawText(
            "运输数量",
            x + colWidths[4] / 2,
            y + ROW_HEIGHT + headerPaint.textSize / 3,
            headerPaint
        )
        x += colWidths[4]

        // "单价"
        canvas.drawRect(x, y, x + colWidths[5], y + ROW_HEIGHT * 2, borderPaint)
        canvas.drawText(
            "单价",
            x + colWidths[5] / 2,
            y + ROW_HEIGHT + headerPaint.textSize / 3,
            headerPaint
        )
        x += colWidths[5]

        // "运输金额"
        canvas.drawRect(x, y, x + colWidths[6], y + ROW_HEIGHT * 2, borderPaint)
        canvas.drawText(
            "运输金额",
            x + colWidths[6] / 2,
            y + ROW_HEIGHT + headerPaint.textSize / 3,
            headerPaint
        )

        y += ROW_HEIGHT

        // Header row 2: "月", "日", (empty, empty for 摘要 sub-columns)
        x = startX

        // "月"
        canvas.drawRect(x, y, x + colWidths[0], y + ROW_HEIGHT, borderPaint)
        canvas.drawText(
            "月",
            x + colWidths[0] / 2,
            y + ROW_HEIGHT / 2 + headerPaint.textSize / 3,
            headerPaint
        )
        x += colWidths[0]

        // "日"
        canvas.drawRect(x, y, x + colWidths[1], y + ROW_HEIGHT, borderPaint)
        canvas.drawText(
            "日",
            x + colWidths[1] / 2,
            y + ROW_HEIGHT / 2 + headerPaint.textSize / 3,
            headerPaint
        )
        x += colWidths[1]

        // Two empty cells under "摘要"
        canvas.drawRect(x, y, x + colWidths[2], y + ROW_HEIGHT, borderPaint)
        x += colWidths[2]
        canvas.drawRect(x, y, x + colWidths[3], y + ROW_HEIGHT, borderPaint)

        y += ROW_HEIGHT

        // Data rows
        for (record in records) {
            x = startX

            // Month
            canvas.drawRect(x, y, x + colWidths[0], y + ROW_HEIGHT, borderPaint)
            canvas.drawText(
                table.month.toString(),
                x + colWidths[0] / 2,
                y + ROW_HEIGHT / 2 + bodyPaint.textSize / 3,
                bodyCenterPaint
            )
            x += colWidths[0]

            // Day
            canvas.drawRect(x, y, x + colWidths[1], y + ROW_HEIGHT, borderPaint)
            canvas.drawText(
                record.day.toString(),
                x + colWidths[1] / 2,
                y + ROW_HEIGHT / 2 + bodyPaint.textSize / 3,
                bodyCenterPaint
            )
            x += colWidths[1]

            // Location 1
            canvas.drawRect(x, y, x + colWidths[2], y + ROW_HEIGHT, borderPaint)
            canvas.drawText(
                record.location1,
                x + CELL_PADDING,
                y + ROW_HEIGHT / 2 + bodyPaint.textSize / 3,
                bodyPaint
            )
            x += colWidths[2]

            // Location 2
            canvas.drawRect(x, y, x + colWidths[3], y + ROW_HEIGHT, borderPaint)
            canvas.drawText(
                record.location2,
                x + CELL_PADDING,
                y + ROW_HEIGHT / 2 + bodyPaint.textSize / 3,
                bodyPaint
            )
            x += colWidths[3]

            // Quantity
            canvas.drawRect(x, y, x + colWidths[4], y + ROW_HEIGHT, borderPaint)
            canvas.drawText(
                String.format("%.3f", record.quantity),
                x + colWidths[4] - CELL_PADDING,
                y + ROW_HEIGHT / 2 + bodyPaint.textSize / 3,
                bodyRightPaint
            )
            x += colWidths[4]

            // Unit Price
            canvas.drawRect(x, y, x + colWidths[5], y + ROW_HEIGHT, borderPaint)
            record.unitPrice?.let { price ->
                canvas.drawText(
                    price.toString(),
                    x + colWidths[5] / 2,
                    y + ROW_HEIGHT / 2 + bodyPaint.textSize / 3,
                    bodyCenterPaint
                )
            }
            x += colWidths[5]

            // Amount
            canvas.drawRect(x, y, x + colWidths[6], y + ROW_HEIGHT, borderPaint)
            canvas.drawText(
                String.format("%.2f", record.amount),
                x + colWidths[6] - CELL_PADDING,
                y + ROW_HEIGHT / 2 + bodyPaint.textSize / 3,
                bodyRightPaint
            )

            y += ROW_HEIGHT
        }

        // Total row
        x = startX
        val totalRowWidth = colWidths[0] + colWidths[1] + colWidths[2] + colWidths[3] + colWidths[4] + colWidths[5]
        canvas.drawRect(x, y, x + totalRowWidth, y + ROW_HEIGHT, borderPaint)
        x += totalRowWidth

        canvas.drawRect(x, y, x + colWidths[6], y + ROW_HEIGHT, borderPaint)
        val totalAmount = records.sumOf { it.amount.toDouble() }.toFloat()
        canvas.drawText(
            String.format("%.2f", totalAmount),
            x + colWidths[6] - CELL_PADDING,
            y + ROW_HEIGHT / 2 + bodyPaint.textSize / 3,
            bodyRightPaint
        )
    }
}
