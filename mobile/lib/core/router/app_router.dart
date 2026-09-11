import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/account/account_pages.dart';
import '../../features/catalog/catalog_pages.dart';
import '../../features/home/home_page.dart';
import '../../features/shell/app_shell.dart';
import '../auth/auth_controller.dart';

final appRouterProvider = Provider<GoRouter>((ref) {
  final auth = ref.watch(authControllerProvider);
  return GoRouter(
    initialLocation: '/',
    redirect: (context, state) {
      if (auth.isLoading) return null;
      final account = auth.valueOrNull;
      final protected = state.matchedLocation == '/buyer-profile' ||
          state.matchedLocation.startsWith('/vendor-');
      if (protected && account == null) {
        return '/login?returnTo=${Uri.encodeComponent(state.uri.toString())}';
      }
      if (state.matchedLocation.startsWith('/vendor-') &&
          account?.isVendor != true) {
        return '/me';
      }
      if (state.matchedLocation == '/buyer-profile' &&
          account?.role != 'buyer') {
        return '/me';
      }
      if (state.matchedLocation == '/login' && account != null) {
        return state.uri.queryParameters['returnTo'] ?? '/me';
      }
      return null;
    },
    routes: [
      StatefulShellRoute.indexedStack(
          builder: (context, state, shell) => AppShell(navigationShell: shell),
          branches: [
            StatefulShellBranch(routes: [
              GoRoute(path: '/', builder: (_, __) => const HomePage())
            ]),
            StatefulShellBranch(routes: [
              GoRoute(
                  path: '/categories', builder: (_, __) => const CatalogPage())
            ]),
            StatefulShellBranch(routes: [
              GoRoute(path: '/me', builder: (_, __) => const MyPage())
            ]),
          ]),
      GoRoute(
          path: '/search',
          builder: (_, __) => const CatalogPage(searchOnly: true)),
      GoRoute(
          path: '/products/:id',
          builder: (_, state) =>
              ProductDetailPage(state.pathParameters['id']!)),
      GoRoute(path: '/vendors', builder: (_, __) => const CatalogPage()),
      GoRoute(
          path: '/vendors/:id',
          builder: (_, state) => VendorDetailPage(state.pathParameters['id']!)),
      GoRoute(
          path: '/login',
          builder: (_, state) =>
              LoginPage(returnTo: state.uri.queryParameters['returnTo'])),
      GoRoute(
          path: '/vendor-center', builder: (_, __) => const VendorCenterPage()),
      GoRoute(
          path: '/vendor-products',
          builder: (_, __) => const VendorProductsPage()),
      GoRoute(
          path: '/buyer-profile', builder: (_, __) => const BuyerProfilePage()),
    ],
  );
});
