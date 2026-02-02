import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../../../theme/context_ext.dart';
import '../../../application/stores/breakpoints_store.dart';
import '../../../domain/entities/intercept_config.dart';
import '../../../domain/entities/intercept_rule.dart';
import 'intercept_rule_card.dart';
import 'intercept_config_form.dart';

class RulesPanel extends StatefulWidget {
  const RulesPanel({super.key});

  @override
  State<RulesPanel> createState() => _RulesPanelState();
}

class _RulesPanelState extends State<RulesPanel> {
  InterceptConfig? _draftConfig;
  List<InterceptRule>? _draftRules;
  bool _savingCfg = false;
  bool _savingRules = false;
  bool _initScheduled = false;

  void _scheduleInitFromStore(BreakpointsStore bp) {
    if (_initScheduled) return;
    if (_draftConfig != null || _draftRules != null) return;
    if (bp.config == null && bp.rules.isEmpty) return;

    _initScheduled = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      setState(() {
        _draftConfig = bp.config;
        _draftRules = _cloneRules(bp.rules);
      });
    });
  }

  List<InterceptRule> _cloneRules(List<InterceptRule> src) {
    return src
        .map(
          (e) => InterceptRule(
            id: e.id,
            enabled: e.enabled,
            priority: e.priority,
            action: e.action,
            once: e.once,
            stopProcessing: e.stopProcessing,
            when: e.when,
          ),
        )
        .toList(growable: true);
  }

  int _nextPriority(List<InterceptRule> rules) {
    var max = -1;
    for (final r in rules) {
      if (r.priority > max) max = r.priority;
    }
    return max + 1;
  }

  void _addRule() {
    final rules = _draftRules;
    if (rules == null) return;
    setState(() {
      rules.add(
        InterceptRule(
          id: '',
          enabled: true,
          priority: _nextPriority(rules),
          action: 'both',
          once: false,
          stopProcessing: true,
          when: const InterceptWhen(),
        ),
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    final bp = context.watch<BreakpointsStore>();
    _scheduleInitFromStore(bp);

    final cfg = _draftConfig ?? bp.config;
    final rules =
        _draftRules ?? (bp.rules.isEmpty ? null : _cloneRules(bp.rules));

    return SingleChildScrollView(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Config', style: context.appText.subtitle),
          const SizedBox(height: 8),
          if (bp.loading && cfg == null)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 12),
              child: LinearProgressIndicator(),
            ),
          if (cfg != null)
            InterceptConfigForm(
              cfg: cfg,
              onChanged: (c) => setState(() => _draftConfig = c),
            ),
          const SizedBox(height: 16),
          Row(
            children: [
              ElevatedButton(
                onPressed: _savingCfg || cfg == null
                    ? null
                    : () async {
                        setState(() => _savingCfg = true);
                        try {
                          await bp.saveConfig(cfg);
                          if (!mounted) return;
                          setState(() => _draftConfig = bp.config);
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(content: Text('Config saved')),
                          );
                        } catch (e) {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(
                              content: Text('Failed to save config: $e'),
                            ),
                          );
                        } finally {
                          if (mounted) setState(() => _savingCfg = false);
                        }
                      },
                child: Text(_savingCfg ? 'Saving…' : 'Save Config'),
              ),
            ],
          ),
          const SizedBox(height: 24),
          Text('Rules', style: context.appText.subtitle),
          const SizedBox(height: 8),
          if (rules == null)
            Text(
              bp.loading ? 'Loading rules…' : 'No rules',
              style: context.appText.body.copyWith(
                color: context.appColors.textSecondary,
              ),
            ),
          if (rules != null)
            ListView.separated(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: rules.length,
              separatorBuilder: (_, __) => const SizedBox(height: 12),
              itemBuilder: (ctx, i) {
                final r = rules[i];
                return InterceptRuleCard(
                  key: ValueKey('rule-$i-${r.id}'),
                  rule: r,
                  onChanged: (nr) {
                    setState(() {
                      rules[i] = nr;
                      _draftRules = rules;
                    });
                  },
                  onDelete: () {
                    setState(() {
                      rules.removeAt(i);
                      _draftRules = rules;
                    });
                  },
                );
              },
            ),
          const SizedBox(height: 8),
          Row(
            children: [
              OutlinedButton.icon(
                onPressed: rules == null ? null : _addRule,
                icon: const Icon(Icons.add),
                label: const Text('Add Rule'),
              ),
              const SizedBox(width: 8),
              ElevatedButton(
                onPressed: _savingRules || rules == null
                    ? null
                    : () async {
                        setState(() => _savingRules = true);
                        try {
                          await bp.replaceRules(rules);
                          await bp.load();
                          if (!mounted) return;
                          setState(() {
                            _draftRules = _cloneRules(bp.rules);
                            _draftConfig ??= bp.config;
                          });
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(content: Text('Rules saved')),
                          );
                        } catch (e) {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(content: Text('Failed to save rules: $e')),
                          );
                        } finally {
                          if (mounted) setState(() => _savingRules = false);
                        }
                      },
                child: Text(_savingRules ? 'Saving…' : 'Save Rules'),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
