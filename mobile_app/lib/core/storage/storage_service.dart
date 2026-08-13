import 'package:shared_preferences/shared_preferences.dart';

class StorageService {
  final SharedPreferences _prefs;

  StorageService(this._prefs);

  static const String _apiKeyKey = 'api_key';
  static const String _subdomainKey = 'subdomain';

  Future<void> setApiKey(String apiKey) async {
    await _prefs.setString(_apiKeyKey, apiKey);
  }

  String? getApiKey() {
    return _prefs.getString(_apiKeyKey);
  }

  Future<void> setSubdomain(String subdomain) async {
    await _prefs.setString(_subdomainKey, subdomain);
  }

  String? getSubdomain() {
    return _prefs.getString(_subdomainKey);
  }

  Future<void> clearAll() async {
    await _prefs.clear();
  }
}
