import 'package:dio/dio.dart';
import '../storage/storage_service.dart';

class DioClient {
  final Dio _dio;
  final StorageService _storageService;

  // Assuming the backend is running locally, we can point to superadmin.localhost:8090/api
  // In production, this would be an environment variable
  static const String baseUrl = 'http://superadmin.localhost:8090/api';

  DioClient(this._storageService) : _dio = Dio(BaseOptions(baseUrl: baseUrl)) {
    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          final apiKey = _storageService.getApiKey();
          if (apiKey != null && apiKey.isNotEmpty) {
            options.headers['X-API-Key'] = apiKey;
          }
          return handler.next(options);
        },
      ),
    );
  }

  Dio get dio => _dio;
}
