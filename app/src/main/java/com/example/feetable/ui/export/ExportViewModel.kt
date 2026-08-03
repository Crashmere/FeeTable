package com.example.feetable.ui.export

import android.app.Application
import android.content.ContentValues
import android.content.Context
import android.content.Intent
import android.graphics.Bitmap
import android.os.Build
import android.os.Environment
import android.provider.MediaStore
import android.widget.Toast
import androidx.core.content.FileProvider
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.feetable.FeeTableApplication
import com.example.feetable.data.entity.FeeRecord
import com.example.feetable.data.entity.FeeTable
import com.example.feetable.export.ExcelExporter
import com.example.feetable.export.PdfExporter
import com.example.feetable.export.PngExporter
import com.example.feetable.export.TableRenderer
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.io.File

class ExportViewModel(application: Application) : AndroidViewModel(application) {
    private val repository = (application as FeeTableApplication).repository

    private val _table = MutableStateFlow<FeeTable?>(null)
    val table: StateFlow<FeeTable?> = _table.asStateFlow()

    private val _records = MutableStateFlow<List<FeeRecord>>(emptyList())
    val records: StateFlow<List<FeeRecord>> = _records.asStateFlow()

    private val _previewBitmap = MutableStateFlow<Bitmap?>(null)
    val previewBitmap: StateFlow<Bitmap?> = _previewBitmap.asStateFlow()

    private val _isExporting = MutableStateFlow(false)
    val isExporting: StateFlow<Boolean> = _isExporting.asStateFlow()

    private var _filterTag: String? = null
    val filterTag: String? get() = _filterTag

    fun loadData(tableId: Int, filterTag: String? = null) {
        _filterTag = filterTag
        viewModelScope.launch {
            _table.value = repository.getTableById(tableId)
            val allRecords = repository.getRecordsByTableIdOnce(tableId)
            _records.value = if (filterTag != null) {
                allRecords.filter { it.tag == filterTag }
            } else {
                allRecords
            }
            generatePreview()
        }
    }

    private fun generatePreview() {
        val t = _table.value ?: return
        val r = _records.value
        if (r.isEmpty()) return

        viewModelScope.launch(Dispatchers.Default) {
            val renderer = TableRenderer(t, r)
            _previewBitmap.value = renderer.render()
        }
    }

    fun exportPng(context: Context) {
        val t = _table.value ?: return
        val r = _records.value
        _isExporting.value = true

        viewModelScope.launch(Dispatchers.IO) {
            try {
                val file = PngExporter(context, t, r).export()
                withContext(Dispatchers.Main) {
                    shareFile(context, file, "image/png")
                    Toast.makeText(context, "图片已生成", Toast.LENGTH_SHORT).show()
                }
            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    Toast.makeText(context, "导出失败: ${e.message}", Toast.LENGTH_SHORT).show()
                }
            } finally {
                _isExporting.value = false
            }
        }
    }

    fun exportPdf(context: Context) {
        val t = _table.value ?: return
        val r = _records.value
        _isExporting.value = true

        viewModelScope.launch(Dispatchers.IO) {
            try {
                val file = PdfExporter(context, t, r).export()
                withContext(Dispatchers.Main) {
                    shareFile(context, file, "application/pdf")
                    Toast.makeText(context, "PDF已生成", Toast.LENGTH_SHORT).show()
                }
            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    Toast.makeText(context, "导出失败: ${e.message}", Toast.LENGTH_SHORT).show()
                }
            } finally {
                _isExporting.value = false
            }
        }
    }

    fun exportExcel(context: Context) {
        val t = _table.value ?: return
        val r = _records.value
        _isExporting.value = true

        viewModelScope.launch(Dispatchers.IO) {
            try {
                val file = ExcelExporter(context, t, r).export()
                withContext(Dispatchers.Main) {
                    shareFile(context, file, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
                    Toast.makeText(context, "Excel已生成", Toast.LENGTH_SHORT).show()
                }
            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    Toast.makeText(context, "导出失败: ${e.message}", Toast.LENGTH_SHORT).show()
                }
            } finally {
                _isExporting.value = false
            }
        }
    }

    fun savePngToGallery(context: Context) {
        val t = _table.value ?: return
        val r = _records.value
        _isExporting.value = true

        viewModelScope.launch(Dispatchers.IO) {
            try {
                val renderer = TableRenderer(t, r)
                val bitmap = renderer.render()
                val tagSuffix = _filterTag?.let { "_$it" } ?: ""
                val filename = "运费明细表_${t.year}年${t.month}月${tagSuffix}.png"

                val contentValues = ContentValues().apply {
                    put(MediaStore.Images.Media.DISPLAY_NAME, filename)
                    put(MediaStore.Images.Media.MIME_TYPE, "image/png")
                    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                        put(MediaStore.Images.Media.RELATIVE_PATH, Environment.DIRECTORY_PICTURES + "/运费明细表")
                        put(MediaStore.Images.Media.IS_PENDING, 1)
                    }
                }

                val resolver = context.contentResolver
                val uri = resolver.insert(MediaStore.Images.Media.EXTERNAL_CONTENT_URI, contentValues)

                uri?.let {
                    resolver.openOutputStream(it)?.use { out ->
                        bitmap.compress(Bitmap.CompressFormat.PNG, 100, out)
                    }
                    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                        contentValues.clear()
                        contentValues.put(MediaStore.Images.Media.IS_PENDING, 0)
                        resolver.update(it, contentValues, null, null)
                    }
                }
                bitmap.recycle()

                withContext(Dispatchers.Main) {
                    Toast.makeText(context, "已保存到相册", Toast.LENGTH_SHORT).show()
                }
            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    Toast.makeText(context, "保存失败: ${e.message}", Toast.LENGTH_SHORT).show()
                }
            } finally {
                _isExporting.value = false
            }
        }
    }

    private fun shareFile(context: Context, file: File, mimeType: String) {
        val uri = FileProvider.getUriForFile(
            context,
            "${context.packageName}.fileprovider",
            file
        )
        val intent = Intent(Intent.ACTION_SEND).apply {
            type = mimeType
            putExtra(Intent.EXTRA_STREAM, uri)
            addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
        }
        context.startActivity(Intent.createChooser(intent, "分享文件"))
    }
}
