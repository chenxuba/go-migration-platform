package com.example.assessment_pad_app

import android.app.Activity
import android.content.Intent
import android.graphics.Color
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.view.View
import android.view.WindowInsets
import android.view.WindowInsetsController
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.embedding.android.FlutterActivity
import io.flutter.plugin.common.MethodChannel

class MainActivity : FlutterActivity() {
    private var pendingSaveResult: MethodChannel.Result? = null
    private var pendingSaveBytes: ByteArray? = null
    private val saveDocumentRequestCode = 4201

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        configureEdgeToEdge()
    }

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        MethodChannel(
            flutterEngine.dartExecutor.binaryMessenger,
            "cn.irts.children.assessmentassistant/file_exporter"
        ).setMethodCallHandler { call, result ->
            if (call.method != "saveFile") {
                result.notImplemented()
                return@setMethodCallHandler
            }
            if (pendingSaveResult != null) {
                result.error("export_in_progress", "已有文件正在导出，请稍后再试", null)
                return@setMethodCallHandler
            }
            val arguments = call.arguments as? Map<*, *>
            val fileName = normalizeDocxName(arguments?.get("fileName") as? String)
            val mimeType = normalizeMimeType(arguments?.get("mimeType") as? String)
            val bytes = arguments?.get("bytes") as? ByteArray
            if (bytes == null || bytes.isEmpty()) {
                result.error("invalid_arguments", "导出文件参数不正确", null)
                return@setMethodCallHandler
            }

            pendingSaveResult = result
            pendingSaveBytes = bytes
            try {
                val intent = Intent(Intent.ACTION_CREATE_DOCUMENT).apply {
                    addCategory(Intent.CATEGORY_OPENABLE)
                    type = mimeType
                    putExtra(Intent.EXTRA_TITLE, fileName)
                }
                startActivityForResult(intent, saveDocumentRequestCode)
            } catch (error: Exception) {
                pendingSaveResult = null
                pendingSaveBytes = null
                result.error("export_failed", "无法打开文件保存器：${error.message}", null)
            }
        }
    }

    override fun onPostResume() {
        super.onPostResume()
        configureEdgeToEdge()
    }

    @Deprecated("Deprecated in Java")
    override fun onActivityResult(requestCode: Int, resultCode: Int, data: Intent?) {
        super.onActivityResult(requestCode, resultCode, data)
        if (requestCode != saveDocumentRequestCode) {
            return
        }
        val result = pendingSaveResult
        val bytes = pendingSaveBytes
        pendingSaveResult = null
        pendingSaveBytes = null

        if (result == null) {
            return
        }
        if (resultCode != Activity.RESULT_OK) {
            result.success(false)
            return
        }
        val uri: Uri? = data?.data
        if (uri == null || bytes == null) {
            result.error("export_failed", "未获取到保存位置", null)
            return
        }
        try {
            contentResolver.openOutputStream(uri)?.use { outputStream ->
                outputStream.write(bytes)
                outputStream.flush()
            } ?: run {
                result.error("export_failed", "无法写入选择的保存位置", null)
                return
            }
            result.success(true)
        } catch (error: Exception) {
            result.error("export_failed", "导出文件写入失败：${error.message}", null)
        }
    }

    private fun configureEdgeToEdge() {
        window.statusBarColor = Color.TRANSPARENT
        window.navigationBarColor = Color.TRANSPARENT

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
            window.setDecorFitsSystemWindows(false)
            window.insetsController?.setSystemBarsAppearance(
                WindowInsetsController.APPEARANCE_LIGHT_STATUS_BARS or
                    WindowInsetsController.APPEARANCE_LIGHT_NAVIGATION_BARS,
                WindowInsetsController.APPEARANCE_LIGHT_STATUS_BARS or
                    WindowInsetsController.APPEARANCE_LIGHT_NAVIGATION_BARS
            )
            window.insetsController?.systemBarsBehavior =
                WindowInsetsController.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
            window.insetsController?.hide(WindowInsets.Type.systemBars())
            return
        }

        @Suppress("DEPRECATION")
        var flags = View.SYSTEM_UI_FLAG_LAYOUT_STABLE or
            View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN or
            View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION or
            View.SYSTEM_UI_FLAG_FULLSCREEN or
            View.SYSTEM_UI_FLAG_HIDE_NAVIGATION or
            View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            @Suppress("DEPRECATION")
            flags = flags or View.SYSTEM_UI_FLAG_LIGHT_STATUS_BAR
        }
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            @Suppress("DEPRECATION")
            flags = flags or View.SYSTEM_UI_FLAG_LIGHT_NAVIGATION_BAR
        }

        @Suppress("DEPRECATION")
        window.decorView.systemUiVisibility = flags
    }

    private fun normalizeMimeType(raw: String?): String {
        val mimeType = raw?.substringBefore(";")?.trim().orEmpty()
        return mimeType.ifEmpty {
            "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
        }
    }

    private fun normalizeDocxName(raw: String?): String {
        val cleaned = raw
            ?.trim()
            ?.replace(Regex("[\\\\/:*?\"<>|]"), "_")
            ?.replace(Regex("\\s+"), " ")
            .orEmpty()
        val name = cleaned.ifEmpty { "IEP计划.docx" }
        return if (name.lowercase().endsWith(".docx")) name else "$name.docx"
    }
}
