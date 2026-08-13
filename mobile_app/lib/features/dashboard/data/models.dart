class DashboardStats {
  final String tenantName;
  final String subdomain;
  final String state;
  final String? licenseExpiryDate;
  final double totalSalesAllTime;
  final int orderCountAllTime;
  final double totalSalesToday;
  final int orderCountToday;
  final double totalSalesYesterday;
  final int orderCountYesterday;
  final double totalSalesMonth;
  final int orderCountMonth;

  DashboardStats({
    required this.tenantName,
    required this.subdomain,
    required this.state,
    this.licenseExpiryDate,
    required this.totalSalesAllTime,
    required this.orderCountAllTime,
    required this.totalSalesToday,
    required this.orderCountToday,
    required this.totalSalesYesterday,
    required this.orderCountYesterday,
    required this.totalSalesMonth,
    required this.orderCountMonth,
  });

  factory DashboardStats.fromJson(Map<String, dynamic> json) {
    return DashboardStats(
      tenantName: json['tenant_name'] ?? '',
      subdomain: json['subdomain'] ?? '',
      state: json['state'] ?? '',
      licenseExpiryDate: json['license_expiry_date'],
      totalSalesAllTime: (json['total_sales_all_time'] ?? 0).toDouble(),
      orderCountAllTime: json['order_count_all_time'] ?? 0,
      totalSalesToday: (json['total_sales_today'] ?? 0).toDouble(),
      orderCountToday: json['order_count_today'] ?? 0,
      totalSalesYesterday: (json['total_sales_yesterday'] ?? 0).toDouble(),
      orderCountYesterday: json['order_count_yesterday'] ?? 0,
      totalSalesMonth: (json['total_sales_month'] ?? 0).toDouble(),
      orderCountMonth: json['order_count_month'] ?? 0,
    );
  }
}

class WeeklyTrend {
  final String date;
  final double totalRevenue;
  final int orderCount;

  WeeklyTrend({
    required this.date,
    required this.totalRevenue,
    required this.orderCount,
  });

  factory WeeklyTrend.fromJson(Map<String, dynamic> json) {
    return WeeklyTrend(
      date: json['date'] ?? '',
      totalRevenue: (json['total_revenue'] ?? 0).toDouble(),
      orderCount: json['order_count'] ?? 0,
    );
  }
}

class RecentOrder {
  final String name;
  final String dateOrder;
  final double amountTotal;
  final String state;

  RecentOrder({
    required this.name,
    required this.dateOrder,
    required this.amountTotal,
    required this.state,
  });

  factory RecentOrder.fromJson(Map<String, dynamic> json) {
    return RecentOrder(
      name: json['name'] ?? '',
      dateOrder: json['date_order'] ?? '',
      amountTotal: (json['amount_total'] ?? 0).toDouble(),
      state: json['state'] ?? '',
    );
  }
}

class TopProduct {
  final String name;
  final int quantity;
  final double totalRevenue;

  TopProduct({
    required this.name,
    required this.quantity,
    required this.totalRevenue,
  });

  factory TopProduct.fromJson(Map<String, dynamic> json) {
    return TopProduct(
      name: json['name'] ?? '',
      quantity: (json['quantity'] ?? 0).toInt(),
      totalRevenue: (json['total_revenue'] ?? 0).toDouble(),
    );
  }
}

class DashboardResponse {
  final DashboardStats stats;
  final List<WeeklyTrend> weeklyTrend;
  final List<RecentOrder> recentOrders;
  final List<TopProduct> topProducts;

  DashboardResponse({
    required this.stats,
    required this.weeklyTrend,
    required this.recentOrders,
    required this.topProducts,
  });

  factory DashboardResponse.fromJson(Map<String, dynamic> json) {
    return DashboardResponse(
      stats: DashboardStats.fromJson(json['stats'] ?? {}),
      weeklyTrend:
          (json['weekly_trend'] as List<dynamic>?)
              ?.map((e) => WeeklyTrend.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      recentOrders:
          (json['recent_orders'] as List<dynamic>?)
              ?.map((e) => RecentOrder.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      topProducts:
          (json['top_products'] as List<dynamic>?)
              ?.map((e) => TopProduct.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}
