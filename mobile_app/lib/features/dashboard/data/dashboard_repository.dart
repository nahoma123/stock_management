import '../../../core/network/dio_client.dart';
import 'models.dart';

class DashboardRepository {
  final DioClient _dioClient;

  DashboardRepository(this._dioClient);

  Future<DashboardResponse> getStats() async {
    try {
      final response = await _dioClient.dio.get('/mobile/stats');
      if (response.statusCode == 200) {
        return DashboardResponse.fromJson(response.data);
      } else {
        throw Exception('Failed to load stats: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Failed to load stats: $e');
    }
  }
}
