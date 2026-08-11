import 'package:dlc_mobile/core/models/models.dart';
import 'package:dlc_mobile/features/market/market_pages.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

void main() {
  testWidgets('market card labels demand information without contact data',
      (tester) async {
    final post = MarketPost(
        id: 1,
        type: 'demand',
        title: '求购液压油泵',
        description: '需要批量采购液压油泵产品',
        publisherName: '采购商',
        status: 'published',
        images: const [],
        expiresAt: DateTime(2026, 9, 1),
        createdAt: DateTime(2026, 8, 1),
        province: '河北',
        quantity: '100 件');
    final router = GoRouter(routes: [
      GoRoute(
          path: '/', builder: (_, __) => Scaffold(body: MarketPostCard(post))),
      GoRoute(path: '/market/:id', builder: (_, __) => const SizedBox())
    ]);
    await tester.pumpWidget(MaterialApp.router(routerConfig: router));
    expect(find.text('求购'), findsOneWidget);
    expect(find.text('求购液压油泵'), findsOneWidget);
    expect(find.text('100 件'), findsOneWidget);
    expect(find.textContaining('电话'), findsNothing);
  });
}
