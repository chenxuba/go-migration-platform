import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'assessment_draft_resume_dialog.dart';
import 'assessment_age_formatter.dart';
import 'assessment_scale_client.dart';
import 'autismdev_assessment_client.dart';
import 'pad_responsive.dart';
import 'pad_top_message.dart';

part 'autismdev_assessment_chrome.dart';
part 'autismdev_assessment_navigation.dart';
part 'autismdev_assessment_workspace.dart';
part 'autismdev_assessment_right_rail.dart';
part 'autismdev_assessment_footer.dart';
part 'autismdev_assessment_detail.dart';
part 'autismdev_assessment_preferences.dart';
part 'autismdev_assessment_loading.dart';
part 'autismdev_assessment_state_actions.dart';
part 'autismdev_assessment_support.dart';

class AutismDevAssessmentPage extends StatefulWidget {
  const AutismDevAssessmentPage({
    required this.onBack,
    this.args = const AutismDevAssessmentLaunchArgs(),
    this.client = const ApiAutismDevAssessmentClient(),
    super.key,
  });

  final VoidCallback onBack;
  final AutismDevAssessmentLaunchArgs args;
  final AutismDevAssessmentClient client;

  @override
  State<AutismDevAssessmentPage> createState() =>
      _AutismDevAssessmentPageState();
}

enum _AutismDevQuestionDisplayPreference { all, matchingAge, ageAndBelow }

class _AutismDevAssessmentPageState extends State<AutismDevAssessmentPage> {
  static const String _authTokenStorageKey = 'auth_token';

  final Map<int, String> _itemScores = <int, String>{};
  final Map<int, String> _itemRemarks = <int, String>{};
  final Map<int, AutismDevAssessmentItem> _itemDetailCache =
      <int, AutismDevAssessmentItem>{};
  final Map<int, Future<AutismDevAssessmentItem>> _itemDetailFetches =
      <int, Future<AutismDevAssessmentItem>>{};
  final TextEditingController _remarkController = TextEditingController();
  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();
  Future<AutismDevDraftDetail?>? _saveDraftFuture;

  AutismDevTemplateSummary _template = AutismDevTemplateSummary.empty;
  AssessmentDraftSummary? _detectedDraft;
  Future<AutismDevDraftDetail>? _detectedDraftDetailRequest;
  String _token = '';
  String _studentName = '';
  String _studentAge = '';
  String _birthDate = '';
  String _assessmentDate = '';
  String _examinerName = '';
  String _selectedDomainCode = '';
  String _selectedRangeFilter = '';
  int _selectedItemNo = 0;
  int _studentId = 0;
  int _draftId = 0;
  int _detectedDraftDetailDraftId = 0;
  int _draftDetectionSerial = 0;
  bool _draftDialogShown = false;
  bool _loading = true;
  bool _saving = false;
  bool _submitting = false;
  bool _autoNext = true;
  bool _saveDraftFutureSilent = false;
  bool _saveDraftJoinedByManual = false;
  _AutismDevQuestionDisplayPreference _questionDisplayPreference =
      _AutismDevQuestionDisplayPreference.ageAndBelow;
  String _errorMessage = '';
  String _autoSaveText = '等待作答';

  @override
  void initState() {
    super.initState();
    _studentId = widget.args.studentId;
    _studentName = widget.args.studentName;
    _studentAge = widget.args.studentAge;
    _birthDate = _dateOnlyText(widget.args.birthDate);
    _assessmentDate = _dateOnlyText(widget.args.assessmentDate).isNotEmpty
        ? _dateOnlyText(widget.args.assessmentDate)
        : _todayIsoDate();
    _examinerName = widget.args.examinerName;
    _draftId = widget.args.draftId;
    _initialize();
  }

  @override
  void dispose() {
    _remarkController.dispose();
    _messageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: _AutismDevColors.page,
      child: Column(
        children: <Widget>[
          _AutismDevTopBar(
            title: _autismDevScaleTitle(widget.args.scaleName),
            studentName: _studentName.trim().isEmpty ? '-' : _studentName,
            studentAge: _studentAgeText,
            assessmentDate: _assessmentDate.isEmpty ? '未设置日期' : _assessmentDate,
            examinerName: _examinerName.trim().isEmpty ? '-' : _examinerName,
            autoSaveText: _autoSaveText,
            saving: _saving,
            submitting: _submitting,
            onBack: widget.onBack,
            onSave: _saveDraft,
            onSubmit: _submitDraft,
          ),
          Expanded(child: _buildBody()),
        ],
      ),
    );
  }

  String get _studentAgeText {
    final String fallback =
        _studentAge.trim().isEmpty ? '未知' : _studentAge.trim();
    return formatAssessmentAgeText(
      birthDate: _birthDate,
      assessmentDate: _assessmentDate,
      fallback: fallback,
    );
  }
}
