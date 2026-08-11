import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../core/theme/app_theme.dart';

class AppShell extends StatelessWidget {
  const AppShell({required this.navigationShell, super.key});
  final StatefulNavigationShell navigationShell;

  @override
  Widget build(BuildContext context) => Scaffold(
        body: navigationShell,
        bottomNavigationBar: NavigationBar(
          selectedIndex: navigationShell.currentIndex,
          onDestinationSelected: (index) => navigationShell.goBranch(index,
              initialLocation: index == navigationShell.currentIndex),
          destinations: const [
            NavigationDestination(
                icon: Icon(Icons.home_outlined),
                selectedIcon: Icon(Icons.home),
                label: '首页'),
            NavigationDestination(
                icon: Icon(Icons.grid_view_outlined),
                selectedIcon: Icon(Icons.grid_view),
                label: '分类'),
            NavigationDestination(icon: _PublishIcon(), label: '发布'),
            NavigationDestination(
                icon: Icon(Icons.assignment_outlined),
                selectedIcon: Icon(Icons.assignment),
                label: '供求'),
            NavigationDestination(
                icon: Icon(Icons.person_outline),
                selectedIcon: Icon(Icons.person),
                label: '我的'),
          ],
        ),
      );
}

class _PublishIcon extends StatelessWidget {
  const _PublishIcon();
  @override
  Widget build(BuildContext context) => Container(
      decoration:
          const BoxDecoration(color: AppTheme.orange, shape: BoxShape.circle),
      padding: const EdgeInsets.all(9),
      child: const Icon(Icons.add, color: Colors.white));
}
