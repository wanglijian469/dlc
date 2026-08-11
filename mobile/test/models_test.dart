import 'package:dlc_mobile/core/models/models.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('market post parser keeps public fields and media', () {
    final post = MarketPost.fromJson({
      'id': 7,
      'type': 'demand',
      'title': '求购收割机链条',
      'description': '长期采购收割机链条和相关配件',
      'publisherName': '采购商',
      'status': 'published',
      'images': ['/api/media/9'],
      'expiresAt': '2026-09-10T00:00:00+08:00',
      'createdAt': '2026-08-10T00:00:00+08:00',
    });
    expect(post.isDemand, isTrue);
    expect(post.images, ['/api/media/9']);
    expect(post.title, contains('链条'));
  });

  test('account role helpers isolate buyer and vendor capabilities', () {
    expect(const Account(username: 'buyer', role: 'buyer').isBuyer, isTrue);
    expect(const Account(username: 'vendor', role: 'vendor').isVendor, isTrue);
    expect(const Account(username: 'admin', role: 'admin').isVendor, isFalse);
  });
}
