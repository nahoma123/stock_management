import '../../../core/network/dio_client.dart';

class SettingsRepository {
  final DioClient _dioClient;

  SettingsRepository(this._dioClient);

  Future<void> registerDevice(String token, String platform) async {
    try {
      final response = await _dioClient.dio.post(
        '/mobile/devices',
        data: {'device_token': token, 'platform': platform},
      );
      if (response.statusCode != 200) {
        throw Exception('Failed to register device: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Failed to register device: $e');
    }
  }

  Future<void> unregisterDevice(String token) async {
    try {
      final response = await _dioClient.dio.delete('/mobile/devices/$token');
      if (response.statusCode != 200) {
        throw Exception('Failed to unregister device: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Failed to unregister device: $e');
    }
  }

  Future<double> updateNotificationSettings(double minAmount) async {
    try {
      final response = await _dioClient.dio.put(
        '/mobile/settings',
        data: {'min_notification_amount': minAmount},
      );
      if (response.statusCode == 200) {
        return (response.data['min_notification_amount'] ?? 0).toDouble();
      } else {
        throw Exception('Failed to update settings: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Failed to update settings: $e');
    }
  }
}
