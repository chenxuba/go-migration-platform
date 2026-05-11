import 'dart:io';
import 'dart:typed_data';

import 'package:file_selector/file_selector.dart';
import 'package:flutter/services.dart' show MethodChannel, PlatformException;

import 'iep_plan_client.dart';

class DownloadedFileSaver {
  DownloadedFileSaver._();

  static const MethodChannel _channel = MethodChannel(
    'cn.irts.children.assessmentassistant/file_exporter',
  );

  static const String _docxMimeType =
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document';

  static Future<bool> save(IepWordFile file) async {
    final Uint8List bytes = _asUint8List(file.bytes);
    if (bytes.isEmpty) {
      throw const IepPlanApiException('导出文件为空');
    }
    final String fileName = _normalizeDocxName(file.resolvedFileName);
    final String mimeType = _normalizeMimeType(file.contentType);

    if (Platform.isIOS || Platform.isAndroid) {
      try {
        final bool? saved = await _channel.invokeMethod<bool>(
          'saveFile',
          <String, Object>{
            'fileName': fileName,
            'mimeType': mimeType,
            'bytes': bytes,
          },
        );
        return saved ?? false;
      } on PlatformException catch (error) {
        final String message = error.message?.trim() ?? '';
        throw IepPlanApiException(message.isEmpty ? '导出文件保存失败' : message);
      }
    }

    final FileSaveLocation? location = await getSaveLocation(
      acceptedTypeGroups: const <XTypeGroup>[
        XTypeGroup(
          label: 'Word 文档',
          extensions: <String>['docx'],
          mimeTypes: <String>[_docxMimeType],
          uniformTypeIdentifiers: <String>[
            'org.openxmlformats.wordprocessingml.document',
          ],
        ),
      ],
      suggestedName: fileName,
      confirmButtonText: '保存',
    );
    if (location == null) {
      return false;
    }
    await File(_normalizeDocxPath(location.path)).writeAsBytes(
      bytes,
      flush: true,
    );
    return true;
  }

  static Uint8List _asUint8List(List<int> bytes) {
    return bytes is Uint8List ? bytes : Uint8List.fromList(bytes);
  }

  static String _normalizeMimeType(String raw) {
    final String mimeType = raw.split(';').first.trim();
    return mimeType.isEmpty ? _docxMimeType : mimeType;
  }

  static String _normalizeDocxName(String raw) {
    final String cleaned = raw
        .replaceAll(RegExp(r'[\\/:*?"<>|]'), '_')
        .replaceAll(RegExp(r'\s+'), ' ')
        .trim();
    final String name = cleaned.isEmpty ? 'IEP计划.docx' : cleaned;
    return name.toLowerCase().endsWith('.docx') ? name : '$name.docx';
  }

  static String _normalizeDocxPath(String path) {
    return path.toLowerCase().endsWith('.docx') ? path : '$path.docx';
  }
}
