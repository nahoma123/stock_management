class TenantInfo {
  final int id;
  final String name;
  final String subdomain;

  TenantInfo({required this.id, required this.name, required this.subdomain});

  factory TenantInfo.fromJson(Map<String, dynamic> json) {
    return TenantInfo(
      id: json['id'] as int,
      name: json['name'] as String,
      subdomain: json['subdomain'] as String,
    );
  }
}
