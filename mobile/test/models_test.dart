import 'package:dlc_mobile/core/models/models.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('account role helpers isolate buyer and vendor capabilities', () {
    expect(const Account(username: 'buyer', role: 'buyer').isBuyer, isTrue);
    expect(const Account(username: 'vendor', role: 'vendor').isVendor, isTrue);
    expect(const Account(username: 'admin', role: 'admin').isVendor, isFalse);
  });
}
