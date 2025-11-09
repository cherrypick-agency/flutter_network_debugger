import 'package:flutter/material.dart';

import '../../domain/entities/performance_overview.dart';

class SlowEndpointsTable extends StatelessWidget {
  const SlowEndpointsTable({required this.endpoints, super.key});

  final List<EndpointPerformance> endpoints;

  @override
  Widget build(BuildContext context) {
    if (endpoints.isEmpty) {
      return const Padding(
        padding: EdgeInsets.all(16),
        child: Text(
          'No endpoints data available',
          style: TextStyle(color: Colors.grey),
        ),
      );
    }

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: DataTable(
        columns: const [
          DataColumn(label: Text('Endpoint')),
          DataColumn(label: Text('Method'), numeric: false),
          DataColumn(label: Text('P95 (ms)'), numeric: true),
          DataColumn(label: Text('Avg (ms)'), numeric: true),
          DataColumn(label: Text('Requests'), numeric: true),
          DataColumn(label: Text('Errors'), numeric: true),
          DataColumn(label: Text('Error Rate'), numeric: true),
        ],
        rows: endpoints.map((ep) {
          return DataRow(
            cells: [
              DataCell(
                SizedBox(
                  width: 300,
                  child: Text(ep.endpoint, overflow: TextOverflow.ellipsis),
                ),
              ),
              DataCell(Text(ep.method)),
              DataCell(Text(ep.p95Duration.toStringAsFixed(1))),
              DataCell(Text(ep.avgDuration.toStringAsFixed(1))),
              DataCell(Text(ep.requestCount.toString())),
              DataCell(
                Text(
                  ep.errorCount.toString(),
                  style: TextStyle(
                    color: ep.errorCount > 0 ? Colors.red : null,
                  ),
                ),
              ),
              DataCell(
                Text(
                  '${(ep.errorRate * 100).toStringAsFixed(1)}%',
                  style: TextStyle(
                    color: ep.errorRate > 0.05 ? Colors.red : Colors.orange,
                  ),
                ),
              ),
            ],
          );
        }).toList(),
      ),
    );
  }
}
