import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/providers.dart';
import '../data/dashboard_repository.dart';
import '../data/models.dart';

final dashboardRepositoryProvider = Provider<DashboardRepository>((ref) {
  final dioClient = ref.watch(dioClientProvider);
  return DashboardRepository(dioClient);
});

final dashboardProvider = FutureProvider.autoDispose<DashboardResponse>((
  ref,
) async {
  final repository = ref.watch(dashboardRepositoryProvider);
  return repository.getStats();
});
