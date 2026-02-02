import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';
import '../../../domain/entities/intercept_config.dart';

class InterceptConfigForm extends StatefulWidget {
  const InterceptConfigForm({
    required this.cfg,
    required this.onChanged,
    super.key,
  });

  final InterceptConfig cfg;
  final ValueChanged<InterceptConfig> onChanged;

  @override
  State<InterceptConfigForm> createState() => _InterceptConfigFormState();
}

class _InterceptConfigFormState extends State<InterceptConfigForm> {
  late final TextEditingController _timeoutCtrl;
  late final TextEditingController _queueMaxCtrl;
  late final TextEditingController _bodyMaxCtrl;
  late final FocusNode _timeoutFocus;
  late final FocusNode _queueFocus;
  late final FocusNode _bodyFocus;
  late int _timeoutLastValid;
  late int _queueLastValid;
  late int _bodyLastValid;

  @override
  void initState() {
    super.initState();
    _timeoutCtrl = TextEditingController(text: widget.cfg.timeoutMs.toString());
    _queueMaxCtrl = TextEditingController(text: widget.cfg.queueMax.toString());
    _bodyMaxCtrl = TextEditingController(
      text: widget.cfg.bodyMaxBytes.toString(),
    );

    _timeoutFocus = FocusNode();
    _queueFocus = FocusNode();
    _bodyFocus = FocusNode();
    _timeoutLastValid = widget.cfg.timeoutMs;
    _queueLastValid = widget.cfg.queueMax;
    _bodyLastValid = widget.cfg.bodyMaxBytes;

    _timeoutFocus.addListener(() {
      if (_timeoutFocus.hasFocus) return;
      final n = _parsePositiveInt(_timeoutCtrl.text);
      if (n != null) {
        _timeoutLastValid = n;
        return;
      }
      _timeoutCtrl.text = _timeoutLastValid.toString();
      _emit(timeoutMs: _timeoutLastValid);
      setState(() {});
    });
    _queueFocus.addListener(() {
      if (_queueFocus.hasFocus) return;
      final n = _parseNonNegativeInt(_queueMaxCtrl.text);
      if (n != null) {
        _queueLastValid = n;
        return;
      }
      _queueMaxCtrl.text = _queueLastValid.toString();
      _emit(queueMax: _queueLastValid);
      setState(() {});
    });
    _bodyFocus.addListener(() {
      if (_bodyFocus.hasFocus) return;
      final n = _parsePositiveInt(_bodyMaxCtrl.text);
      if (n != null) {
        _bodyLastValid = n;
        return;
      }
      _bodyMaxCtrl.text = _bodyLastValid.toString();
      _emit(bodyMaxBytes: _bodyLastValid);
      setState(() {});
    });
  }

  @override
  void didUpdateWidget(covariant InterceptConfigForm oldWidget) {
    super.didUpdateWidget(oldWidget);

    if (_timeoutCtrl.text != widget.cfg.timeoutMs.toString()) {
      _timeoutCtrl.text = widget.cfg.timeoutMs.toString();
      _timeoutLastValid = widget.cfg.timeoutMs;
    }
    if (_queueMaxCtrl.text != widget.cfg.queueMax.toString()) {
      _queueMaxCtrl.text = widget.cfg.queueMax.toString();
      _queueLastValid = widget.cfg.queueMax;
    }
    if (_bodyMaxCtrl.text != widget.cfg.bodyMaxBytes.toString()) {
      _bodyMaxCtrl.text = widget.cfg.bodyMaxBytes.toString();
      _bodyLastValid = widget.cfg.bodyMaxBytes;
    }
  }

  @override
  void dispose() {
    _timeoutCtrl.dispose();
    _queueMaxCtrl.dispose();
    _bodyMaxCtrl.dispose();
    _timeoutFocus.dispose();
    _queueFocus.dispose();
    _bodyFocus.dispose();
    super.dispose();
  }

  void _emit({
    bool? enabled,
    bool? requests,
    bool? responses,
    int? timeoutMs,
    int? queueMax,
    int? bodyMaxBytes,
    bool? reencode,
    String? overflow,
  }) {
    widget.onChanged(
      InterceptConfig(
        enabled: enabled ?? widget.cfg.enabled,
        requests: requests ?? widget.cfg.requests,
        responses: responses ?? widget.cfg.responses,
        timeoutMs: timeoutMs ?? widget.cfg.timeoutMs,
        queueMax: queueMax ?? widget.cfg.queueMax,
        bodyMaxBytes: bodyMaxBytes ?? widget.cfg.bodyMaxBytes,
        reencode: reencode ?? widget.cfg.reencode,
        overflow: overflow ?? widget.cfg.overflow,
      ),
    );
  }

  int? _parsePositiveInt(String v) {
    final n = int.tryParse(v.trim());
    if (n == null || n <= 0) return null;
    return n;
  }

  int? _parseNonNegativeInt(String v) {
    final n = int.tryParse(v.trim());
    if (n == null || n < 0) return null;
    return n;
  }

  @override
  Widget build(BuildContext context) {
    final timeoutErr =
        _timeoutCtrl.text.trim().isEmpty ||
            _parsePositiveInt(_timeoutCtrl.text) != null
        ? null
        : 'Должно быть > 0';
    final queueErr =
        _queueMaxCtrl.text.trim().isEmpty ||
            _parseNonNegativeInt(_queueMaxCtrl.text) != null
        ? null
        : 'Должно быть >= 0';
    final bodyErr =
        _bodyMaxCtrl.text.trim().isEmpty ||
            _parsePositiveInt(_bodyMaxCtrl.text) != null
        ? null
        : 'Должно быть > 0';

    return Wrap(
      spacing: 12,
      runSpacing: 12,
      children: [
        FilterChip(
          selected: widget.cfg.enabled,
          label: const Text('Enabled'),
          onSelected: (v) => _emit(enabled: v),
        ),
        FilterChip(
          selected: widget.cfg.requests,
          label: const Text('Requests'),
          onSelected: (v) => _emit(requests: v),
        ),
        FilterChip(
          selected: widget.cfg.responses,
          label: const Text('Responses'),
          onSelected: (v) => _emit(responses: v),
        ),
        SizedBox(
          width: 160,
          child: TextField(
            controller: _timeoutCtrl,
            focusNode: _timeoutFocus,
            decoration: InputDecoration(
              labelText: 'Timeout (ms)',
              errorText: timeoutErr,
            ),
            keyboardType: TextInputType.number,
            onChanged: (v) {
              final n = _parsePositiveInt(v);
              if (n == null) return;
              _timeoutLastValid = n;
              _emit(timeoutMs: n);
            },
          ),
        ),
        SizedBox(
          width: 160,
          child: TextField(
            controller: _queueMaxCtrl,
            focusNode: _queueFocus,
            decoration: InputDecoration(
              labelText: 'Queue max (0 = unlimited)',
              errorText: queueErr,
            ),
            keyboardType: TextInputType.number,
            onChanged: (v) {
              final n = _parseNonNegativeInt(v);
              if (n == null) return;
              _queueLastValid = n;
              _emit(queueMax: n);
            },
          ),
        ),
        SizedBox(
          width: 180,
          child: TextField(
            controller: _bodyMaxCtrl,
            focusNode: _bodyFocus,
            decoration: InputDecoration(
              labelText: 'Body max bytes',
              errorText: bodyErr,
            ),
            keyboardType: TextInputType.number,
            onChanged: (v) {
              final n = _parsePositiveInt(v);
              if (n == null) return;
              _bodyLastValid = n;
              _emit(bodyMaxBytes: n);
            },
          ),
        ),
        FilterChip(
          selected: widget.cfg.reencode,
          label: const Text('Re-encode bodies'),
          onSelected: (v) => _emit(reencode: v),
        ),
        SizedBox(
          width: 220,
          child: DropdownButtonFormField<String>(
            value: widget.cfg.overflow,
            decoration: const InputDecoration(labelText: 'Overflow policy'),
            onChanged: (v) => _emit(overflow: v ?? widget.cfg.overflow),
            items: const [
              'auto-continue-oldest',
              'drop-new',
            ].map((e) => DropdownMenuItem(value: e, child: Text(e))).toList(),
          ),
        ),
        Padding(
          padding: const EdgeInsets.only(top: 4),
          child: Text(
            'timeout/bodyMaxBytes принимаются бэком только если > 0, queueMax — если >= 0.',
            style: context.appText.body.copyWith(
              color: context.appColors.textSecondary,
            ),
          ),
        ),
      ],
    );
  }
}
