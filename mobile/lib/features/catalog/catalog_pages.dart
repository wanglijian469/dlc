import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/auth/auth_controller.dart';
import '../../core/models/models.dart';
import '../../core/network/api_client.dart';
import '../../core/widgets/remote_image.dart';

class CatalogPage extends ConsumerStatefulWidget {
  const CatalogPage({this.searchOnly = false, super.key});
  final bool searchOnly;
  @override
  ConsumerState<CatalogPage> createState() => _CatalogPageState();
}

class _CatalogPageState extends ConsumerState<CatalogPage>
    with SingleTickerProviderStateMixin {
  late final TabController _tabs = TabController(length: 2, vsync: this);
  final _search = TextEditingController();
  Future<List<Object>>? _future;
  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _tabs.dispose();
    _search.dispose();
    super.dispose();
  }

  void _load() => setState(() => _future = _fetch());
  Future<List<Object>> _fetch() async {
    final api = ref.read(apiClientProvider);
    final keyword = _search.text.trim();
    final values = await Future.wait([
      api.getJson('/api/products', query: {'keyword': keyword, 'pageSize': 40}),
      api.getJson('/api/vendors', query: {'keyword': keyword, 'pageSize': 40})
    ]);
    return [
      ApiClient.page(values[0], ProductSummary.fromJson),
      ApiClient.page(values[1], VendorSummary.fromJson)
    ];
  }

  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(
            title: Text(widget.searchOnly ? '全站搜索' : '产品与厂家'),
            bottom: TabBar(
                controller: _tabs,
                tabs: const [Tab(text: '配件产品'), Tab(text: '厂家目录')])),
        body: Column(children: [
          Padding(
              padding: const EdgeInsets.all(12),
              child: SearchBar(
                  controller: _search,
                  hintText: '输入配件、型号或厂家',
                  leading: const Icon(Icons.search),
                  trailing: [
                    IconButton(
                        onPressed: _load, icon: const Icon(Icons.arrow_forward))
                  ],
                  onSubmitted: (_) => _load())),
          Expanded(
              child: FutureBuilder<List<Object>>(
                  future: _future,
                  builder: (context, snapshot) {
                    if (snapshot.connectionState != ConnectionState.done) {
                      return const Center(child: CircularProgressIndicator());
                    }
                    if (snapshot.hasError) {
                      return Center(
                          child: TextButton(
                              onPressed: _load,
                              child: Text('加载失败：${snapshot.error}\n点击重试',
                                  textAlign: TextAlign.center)));
                    }
                    final products =
                        snapshot.data![0] as PageData<ProductSummary>;
                    final vendors =
                        snapshot.data![1] as PageData<VendorSummary>;
                    return TabBarView(controller: _tabs, children: [
                      _ProductList(products.items),
                      _VendorList(vendors.items)
                    ]);
                  }))
        ]),
      );
}

class _ProductList extends StatelessWidget {
  const _ProductList(this.items);
  final List<ProductSummary> items;
  @override
  Widget build(BuildContext context) => items.isEmpty
      ? const Center(child: Text('暂无匹配产品'))
      : GridView.builder(
          padding: const EdgeInsets.all(12),
          itemCount: items.length,
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 2,
              childAspectRatio: .72,
              crossAxisSpacing: 10,
              mainAxisSpacing: 10),
          itemBuilder: (_, index) {
            final item = items[index];
            return Card(
                clipBehavior: Clip.antiAlias,
                child: InkWell(
                    onTap: () =>
                        context.push('/products/${item.slug ?? item.id}'),
                    child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Expanded(
                              child: RemoteImage(item.image,
                                  width: double.infinity)),
                          Padding(
                              padding: const EdgeInsets.all(10),
                              child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(item.name,
                                        maxLines: 2,
                                        overflow: TextOverflow.ellipsis,
                                        style: const TextStyle(
                                            fontWeight: FontWeight.bold)),
                                    const SizedBox(height: 5),
                                    Text(item.priceNote ?? '价格面议',
                                        style: TextStyle(
                                            color: Theme.of(context)
                                                .colorScheme
                                                .secondary))
                                  ]))
                        ])));
          });
}

class _VendorList extends StatelessWidget {
  const _VendorList(this.items);
  final List<VendorSummary> items;
  @override
  Widget build(BuildContext context) => items.isEmpty
      ? const Center(child: Text('暂无匹配厂家'))
      : ListView.separated(
          padding: const EdgeInsets.all(12),
          itemCount: items.length,
          separatorBuilder: (_, __) => const SizedBox(height: 9),
          itemBuilder: (_, index) {
            final item = items[index];
            return Card(
                child: ListTile(
                    onTap: () =>
                        context.push('/vendors/${item.slug ?? item.id}'),
                    leading: ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: RemoteImage(item.logo, width: 54, height: 54)),
                    title: Text(item.name),
                    subtitle: Text(
                        [item.province, item.city, item.mainProducts]
                            .whereType<String>()
                            .where((value) => value.isNotEmpty)
                            .join(' · '),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis),
                    trailing: const Icon(Icons.chevron_right)));
          });
}

class ProductDetailPage extends ConsumerWidget {
  const ProductDetailPage(this.id, {super.key});
  final String id;
  @override
  Widget build(BuildContext context, WidgetRef ref) => _JsonDetail(
      title: '产品详情',
      future: ref.read(apiClientProvider).getJson(RegExp(r'^\d+$').hasMatch(id)
          ? '/api/products/$id'
          : '/api/products/slug/$id'),
      kind: 'product');
}

class VendorDetailPage extends ConsumerWidget {
  const VendorDetailPage(this.id, {super.key});
  final String id;
  @override
  Widget build(BuildContext context, WidgetRef ref) => _JsonDetail(
      title: '厂家详情',
      future: ref.read(apiClientProvider).getJson(RegExp(r'^\d+$').hasMatch(id)
          ? '/api/vendors/$id'
          : '/api/vendors/slug/$id'),
      kind: 'vendor');
}

class _JsonDetail extends StatelessWidget {
  const _JsonDetail(
      {required this.title, required this.future, required this.kind});
  final String title;
  final Future<Map<String, dynamic>> future;
  final String kind;
  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(title: Text(title)),
        body: FutureBuilder<Map<String, dynamic>>(
          future: future,
          builder: (context, snapshot) {
            if (!snapshot.hasData) {
              return snapshot.hasError
                  ? Center(child: Text('加载失败：${snapshot.error}'))
                  : const Center(child: CircularProgressIndicator());
            }
            final data = snapshot.data!;
            final image = kind == 'vendor'
                ? (data['coverImage'] ?? data['logo']) as String?
                : data['image'] as String?;
            final name = data['name'] as String? ?? title;
            final description = (data['description'] ??
                    data['detailContent'] ??
                    data['mainProducts']) as String? ??
                '暂无详细说明';
            return ListView(
                padding: const EdgeInsets.fromLTRB(14, 14, 14, 90),
                children: [
                  ClipRRect(
                      borderRadius: BorderRadius.circular(14),
                      child: RemoteImage(image,
                          height: 240, width: double.infinity)),
                  const SizedBox(height: 18),
                  Text(name,
                      style: Theme.of(context)
                          .textTheme
                          .headlineSmall
                          ?.copyWith(fontWeight: FontWeight.bold)),
                  const SizedBox(height: 12),
                  if (data['province'] != null)
                    Text('${data['province']} · ${data['city'] ?? ''}'),
                  const SizedBox(height: 18),
                  Text(description, style: const TextStyle(height: 1.7)),
                ]);
          },
        ),
      );
}
