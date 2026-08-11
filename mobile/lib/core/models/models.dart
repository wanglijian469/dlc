class Account {
  const Account({required this.username, required this.role, this.vendorId});
  final String username;
  final String role;
  final int? vendorId;
  bool get isBuyer => role == 'buyer';
  bool get isVendor => role == 'vendor';

  factory Account.fromJson(Map<String, dynamic> json) => Account(
        username: json['username'] as String? ?? '',
        role: json['role'] as String? ?? '',
        vendorId: (json['vendorId'] as num?)?.toInt(),
      );
}

class ProductSummary {
  const ProductSummary(
      {required this.id,
      required this.name,
      this.image,
      this.description,
      this.priceNote,
      this.slug});
  final int id;
  final String name;
  final String? image;
  final String? description;
  final String? priceNote;
  final String? slug;
  factory ProductSummary.fromJson(Map<String, dynamic> json) => ProductSummary(
        id: (json['id'] as num).toInt(),
        name: json['name'] as String? ?? '',
        image: json['image'] as String?,
        description: json['description'] as String?,
        priceNote: json['priceNote'] as String?,
        slug: json['slug'] as String?,
      );
}

class VendorSummary {
  const VendorSummary(
      {required this.id,
      required this.name,
      this.slug,
      this.logo,
      this.province,
      this.city,
      this.mainProducts});
  final int id;
  final String name;
  final String? slug;
  final String? logo;
  final String? province;
  final String? city;
  final String? mainProducts;
  factory VendorSummary.fromJson(Map<String, dynamic> json) => VendorSummary(
        id: (json['id'] as num).toInt(),
        name: json['name'] as String? ?? '',
        slug: json['slug'] as String?,
        logo: json['logo'] as String?,
        province: json['province'] as String?,
        city: json['city'] as String?,
        mainProducts: json['mainProducts'] as String?,
      );
}

class MarketPost {
  const MarketPost(
      {required this.id,
      required this.type,
      required this.title,
      required this.description,
      required this.publisherName,
      required this.status,
      required this.images,
      required this.expiresAt,
      required this.createdAt,
      this.categoryId,
      this.compatibleModels,
      this.province,
      this.city,
      this.quantity,
      this.deliveryNote,
      this.vendorId});
  final int id;
  final String type;
  final String title;
  final String description;
  final String publisherName;
  final String status;
  final List<String> images;
  final DateTime expiresAt;
  final DateTime createdAt;
  final int? categoryId;
  final String? compatibleModels;
  final String? province;
  final String? city;
  final String? quantity;
  final String? deliveryNote;
  final int? vendorId;
  bool get isDemand => type == 'demand';

  factory MarketPost.fromJson(Map<String, dynamic> json) => MarketPost(
        id: (json['id'] as num).toInt(),
        type: json['type'] as String? ?? 'demand',
        title: json['title'] as String? ?? '',
        description: json['description'] as String? ?? '',
        publisherName: json['publisherName'] as String? ?? '',
        status: json['status'] as String? ?? 'published',
        images: (json['images'] as List? ?? const [])
            .map((value) => value.toString())
            .toList(),
        expiresAt: DateTime.parse(json['expiresAt'] as String),
        createdAt: DateTime.parse(json['createdAt'] as String),
        categoryId: (json['categoryId'] as num?)?.toInt(),
        compatibleModels: json['compatibleModels'] as String?,
        province: json['province'] as String?,
        city: json['city'] as String?,
        quantity: json['quantity'] as String?,
        deliveryNote: json['deliveryNote'] as String?,
        vendorId: (json['vendorId'] as num?)?.toInt(),
      );
}

class PageData<T> {
  const PageData(
      {required this.items,
      required this.page,
      required this.pageSize,
      required this.total});
  final List<T> items;
  final int page;
  final int pageSize;
  final int total;
}
