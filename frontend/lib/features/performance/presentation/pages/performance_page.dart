import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';

import '../../../../core/di/di.dart';
import '../../application/stores/performance_store.dart';
import '../widgets/metric_card.dart';
import '../widgets/slow_endpoints_table.dart';

class PerformancePage extends StatefulWidget {
  const PerformancePage({super.key});

  @override
  State<PerformancePage> createState() => _PerformancePageState();
}

class _PerformancePageState extends State<PerformancePage> {
  @override
  void initState() {
    super.initState();
    final store = sl<PerformanceStore>();
    store.loadMetrics();
  }

  @override
  Widget build(BuildContext context) {
    final store = sl<PerformanceStore>();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Performance Insights'),
        actions: [
          PopupMenuButton<String>(
            onSelected: (value) {
              switch (value) {
                case 'hour':
                  store.setLastHour();
                  break;
                case 'day':
                  store.setLastDay();
                  break;
                case 'week':
                  store.setLastWeek();
                  break;
              }
            },
            itemBuilder: (context) => [
              const PopupMenuItem(value: 'hour', child: Text('Last Hour')),
              const PopupMenuItem(value: 'day', child: Text('Last 24 Hours')),
              const PopupMenuItem(value: 'week', child: Text('Last 7 Days')),
            ],
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Row(
                children: [
                  Observer(
                    builder: (_) {
                      final duration = store.to.difference(store.from);
                      String label;
                      if (duration.inHours < 2) {
                        label = 'Last Hour';
                      } else if (duration.inHours < 25) {
                        label = 'Last 24h';
                      } else {
                        label = 'Last 7d';
                      }
                      return Text(label);
                    },
                  ),
                  const Icon(Icons.arrow_drop_down),
                ],
              ),
            ),
          ),
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => store.loadMetrics(),
          ),
        ],
      ),
      body: Observer(
        builder: (_) {
          if (store.loading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (store.error != null) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.error_outline, size: 48, color: Colors.red),
                  const SizedBox(height: 16),
                  Text('Error: ${store.error}'),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => store.loadMetrics(),
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          final overview = store.overview;
          if (overview == null) {
            return const Center(child: Text('No data available'));
          }

          return SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Metric Cards
                GridView.count(
                  crossAxisCount: 4,
                  shrinkWrap: true,
                  physics: const NeverScrollableScrollPhysics(),
                  mainAxisSpacing: 16,
                  crossAxisSpacing: 16,
                  childAspectRatio: 2,
                  children: [
                    MetricCard(
                      title: 'Avg Response Time',
                      value:
                          '${overview.responseTimeStats.mean.toStringAsFixed(1)}ms',
                      icon: Icons.speed,
                      color: Colors.blue,
                    ),
                    MetricCard(
                      title: 'P95 Response Time',
                      value:
                          '${overview.responseTimeStats.p95.toStringAsFixed(1)}ms',
                      icon: Icons.show_chart,
                      color: Colors.purple,
                    ),
                    MetricCard(
                      title: 'Throughput',
                      value:
                          '${overview.throughputStats.requestsPerSecond.toStringAsFixed(2)} req/s',
                      icon: Icons.trending_up,
                      color: Colors.green,
                    ),
                    MetricCard(
                      title: 'Error Rate',
                      value:
                          '${(overview.throughputStats.errorRate * 100).toStringAsFixed(1)}%',
                      icon: Icons.error_outline,
                      color: overview.throughputStats.errorRate > 0.05
                          ? Colors.red
                          : Colors.orange,
                    ),
                  ],
                ),

                const SizedBox(height: 32),

                // Response Time Stats
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Response Time Distribution',
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 16),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceAround,
                          children: [
                            _StatItem(
                              label: 'Min',
                              value:
                                  '${overview.responseTimeStats.min.toStringAsFixed(1)}ms',
                            ),
                            _StatItem(
                              label: 'P50',
                              value:
                                  '${overview.responseTimeStats.p50.toStringAsFixed(1)}ms',
                            ),
                            _StatItem(
                              label: 'P95',
                              value:
                                  '${overview.responseTimeStats.p95.toStringAsFixed(1)}ms',
                            ),
                            _StatItem(
                              label: 'P99',
                              value:
                                  '${overview.responseTimeStats.p99.toStringAsFixed(1)}ms',
                            ),
                            _StatItem(
                              label: 'Max',
                              value:
                                  '${overview.responseTimeStats.max.toStringAsFixed(1)}ms',
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text(
                          'Total Requests: ${overview.responseTimeStats.count}',
                          style: TextStyle(color: Colors.grey.shade600),
                        ),
                      ],
                    ),
                  ),
                ),

                const SizedBox(height: 16),

                // Throughput Stats
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Request Statistics',
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 16),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceAround,
                          children: [
                            _StatItem(
                              label: 'Total',
                              value:
                                  '${overview.throughputStats.totalRequests}',
                            ),
                            _StatItem(
                              label: 'Success (2xx)',
                              value: '${overview.throughputStats.status2xx}',
                              color: Colors.green,
                            ),
                            _StatItem(
                              label: 'Client Errors (4xx)',
                              value: '${overview.throughputStats.status4xx}',
                              color: Colors.orange,
                            ),
                            _StatItem(
                              label: 'Server Errors (5xx)',
                              value: '${overview.throughputStats.status5xx}',
                              color: Colors.red,
                            ),
                            _StatItem(
                              label: 'Success Rate',
                              value:
                                  '${(overview.throughputStats.successRate * 100).toStringAsFixed(1)}%',
                              color: Colors.blue,
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ),

                const SizedBox(height: 16),

                // Top Slow Endpoints
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Top Slow Endpoints',
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 16),
                        SlowEndpointsTable(
                          endpoints: overview.topSlowEndpoints,
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          );
        },
      ),
    );
  }
}

class _StatItem extends StatelessWidget {
  const _StatItem({required this.label, required this.value, this.color});

  final String label;
  final String value;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(
          label,
          style: TextStyle(color: Colors.grey.shade600, fontSize: 12),
        ),
        const SizedBox(height: 4),
        Text(
          value,
          style: TextStyle(
            fontSize: 20,
            fontWeight: FontWeight.bold,
            color: color,
          ),
        ),
      ],
    );
  }
}
