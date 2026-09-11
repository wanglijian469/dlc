import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/auth/auth_controller.dart';
import '../../core/models/models.dart';
import '../../core/network/api_client.dart';
import '../../core/widgets/remote_image.dart';

final homeProvider = FutureProvider<_HomeData>((ref) async {
  final api = ref.watch(apiClientProvider);
  final values = await Future.wait([
    api.getJson('/api/home'),
    api.getJson('/api/products', query: {'recommended': true, 'pageSize': 6}),
  ]);
  final home = values[0];
  return _HomeData(
    products: ApiClient.page(values[1], ProductSummary.fromJson).items,
    vendors: [
      ...(home['recommendedVendors'] as List? ?? const []),
      ...(home['moreVendors'] as List? ?? const [])
    ]
        .map((item) => VendorSummary.fromJson(item as Map<String, dynamic>))
        .fold(
            <VendorSummary>[],
            (result, item) => result.any((vendor) => vendor.id == item.id)
                ? result
                : [...result, item])
        .take(6)
        .toList(),
  );
});

class HomePage extends ConsumerWidget {
  const HomePage({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) => Scaffold(
        appBar: AppBar(title: const Text('大陆农机配件'), actions: [
          IconButton(
              onPressed: () => context.go('/search'),
              icon: const Icon(Icons.search))
        ]),
        body: RefreshIndicator(
          onRefresh: () => ref.refresh(homeProvider.future),
          child: ref.watch(homeProvider).when(
                loading: () => const Center(child: CircularProgressIndicator()),
                error: (error, _) => _ErrorView(
                    message: error.toString(),
                    retry: () => ref.invalidate(homeProvider)),
                data: (data) => ListView(
                    padding: const EdgeInsets.fromLTRB(14, 14, 14, 28),
                    children: [
                      SearchBar(
                          hintText: '搜索配件、型号、厂家',
                          leading: const Icon(Icons.search),
                          onTap: () => context.go('/search')),
                      const SizedBox(height: 16),
                      Row(
                          mainAxisAlignment: MainAxisAlignment.spaceAround,
                          children: [
                            _Shortcut(
                                icon: Icons.grid_view,
                                label: '找配件',
                                onTap: () => context.go('/categories')),
                            _Shortcut(
                                icon: Icons.factory_outlined,
                                label: '找厂家',
                                onTap: () => context.push('/vendors')),
                            _Shortcut(
                                icon: Icons.handyman_outlined,
                                label: '加工服务',
                                onTap: () =>
                                    context.push('/vendors?processing=true')),
                            _Shortcut(
                                icon: Icons.person_outline,
                                label: '我的账号',
                                onTap: () => context.go('/me')),
                          ]),
                      if (data.products.isNotEmpty) ...[
                        _Title('推荐配件', onMore: () => context.go('/categories')),
                        GridView.builder(
                            shrinkWrap: true,
                            physics: const NeverScrollableScrollPhysics(),
                            itemCount: data.products.length,
                            gridDelegate:
                                const SliverGridDelegateWithFixedCrossAxisCount(
                                    crossAxisCount: 2,
                                    childAspectRatio: .74,
                                    crossAxisSpacing: 10,
                                    mainAxisSpacing: 10),
                            itemBuilder: (_, index) =>
                                _ProductTile(data.products[index]))
                      ],
                      if (data.vendors.isNotEmpty) ...[
                        _Title('源头厂家', onMore: () => context.push('/vendors')),
                        ...data.vendors.map((vendor) => Card(
                            child: ListTile(
                                onTap: () => context.push(
                                    '/vendors/${vendor.slug ?? vendor.id}'),
                                leading: ClipRRect(
                                    borderRadius: BorderRadius.circular(8),
                                    child: RemoteImage(vendor.logo,
                                        width: 48, height: 48)),
                                title: Text(vendor.name,
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis),
                                subtitle: Text(
                                    [vendor.province, vendor.mainProducts]
                                        .whereType<String>()
                                        .where((value) => value.isNotEmpty)
                                        .join(' · '),
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis),
                                trailing: const Icon(Icons.chevron_right))))
                      ],
                    ]),
              ),
        ),
      );
}

class _HomeData {
  const _HomeData({required this.products, required this.vendors});
  final List<ProductSummary> products;
  final List<VendorSummary> vendors;
}

class _Shortcut extends StatelessWidget {
  const _Shortcut(
      {required this.icon, required this.label, required this.onTap});
  final IconData icon;
  final String label;
  final VoidCallback onTap;
  @override
  Widget build(BuildContext context) => InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(14),
      child: Padding(
          padding: const EdgeInsets.all(7),
          child: Column(children: [
            CircleAvatar(
                radius: 25,
                backgroundColor: Theme.of(context).colorScheme.primaryContainer,
                child: Icon(icon)),
            const SizedBox(height: 6),
            Text(label, style: Theme.of(context).textTheme.labelMedium)
          ])));
}

class _Title extends StatelessWidget {
  const _Title(this.title, {required this.onMore});
  final String title;
  final VoidCallback onMore;
  @override
  Widget build(BuildContext context) => Padding(
      padding: const EdgeInsets.only(top: 24, bottom: 10),
      child: Row(children: [
        Expanded(
            child: Text(title,
                style: Theme.of(context)
                    .textTheme
                    .titleLarge
                    ?.copyWith(fontWeight: FontWeight.bold))),
        TextButton(onPressed: onMore, child: const Text('查看全部'))
      ]));
}

class _ProductTile extends StatelessWidget {
  const _ProductTile(this.product);
  final ProductSummary product;
  @override
  Widget build(BuildContext context) => Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
          onTap: () => context.push('/products/${product.slug ?? product.id}'),
          child:
              Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Expanded(child: RemoteImage(product.image, width: double.infinity)),
            Padding(
                padding: const EdgeInsets.all(10),
                child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(product.name,
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(fontWeight: FontWeight.bold)),
                      const SizedBox(height: 5),
                      Text(product.priceNote ?? '价格面议',
                          style: TextStyle(
                              color: Theme.of(context).colorScheme.secondary))
                    ]))
          ])));
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.message, required this.retry});
  final String message;
  final VoidCallback retry;
  @override
  Widget build(BuildContext context) => ListView(children: [
        const SizedBox(height: 150),
        const Icon(Icons.cloud_off, size: 44),
        Padding(
            padding: const EdgeInsets.all(16),
            child: Text(message, textAlign: TextAlign.center)),
        Center(
            child:
                FilledButton.tonal(onPressed: retry, child: const Text('重试')))
      ]);
}
