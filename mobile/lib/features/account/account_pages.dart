import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/auth/auth_controller.dart';
import '../../core/models/models.dart';
import '../../core/network/api_client.dart';
import '../market/market_pages.dart';

class LoginPage extends ConsumerStatefulWidget {
  const LoginPage({this.returnTo, super.key});
  final String? returnTo;
  @override
  ConsumerState<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends ConsumerState<LoginPage>
    with SingleTickerProviderStateMixin {
  late final tabs = TabController(length: 2, vsync: this);
  final username = TextEditingController();
  final password = TextEditingController();
  final confirm = TextEditingController();
  final company = TextEditingController();
  final displayName = TextEditingController();
  final contact = TextEditingController();
  final phone = TextEditingController();
  bool register = false;
  String role = 'buyer';
  Future<void> submit() async {
    final controller = ref.read(authControllerProvider.notifier);
    if (register) {
      if (password.text != confirm.text) {
        ScaffoldMessenger.of(context)
            .showSnackBar(const SnackBar(content: Text('两次密码不一致')));
        return;
      }
      await controller.register(
          username: username.text,
          password: password.text,
          role: role,
          companyName: company.text,
          displayName: displayName.text,
          contactName: contact.text,
          phone: phone.text);
    } else {
      await controller.login(username.text, password.text);
    }
    if (mounted && ref.read(authControllerProvider).valueOrNull != null) {
      context.go(
          widget.returnTo?.startsWith('/') == true ? widget.returnTo! : '/me');
    }
  }

  @override
  Widget build(BuildContext context) {
    final auth = ref.watch(authControllerProvider);
    return Scaffold(
        appBar: AppBar(title: const Text('账号中心')),
        body: ListView(padding: const EdgeInsets.all(20), children: [
          const SizedBox(height: 22),
          Icon(Icons.agriculture,
              size: 62, color: Theme.of(context).colorScheme.primary),
          const SizedBox(height: 12),
          Text('连接真实需求与源头厂家',
              textAlign: TextAlign.center,
              style: Theme.of(context)
                  .textTheme
                  .headlineSmall
                  ?.copyWith(fontWeight: FontWeight.bold)),
          const SizedBox(height: 28),
          SegmentedButton<bool>(
              segments: const [
                ButtonSegment(value: false, label: Text('登录')),
                ButtonSegment(value: true, label: Text('注册'))
              ],
              selected: {
                register
              },
              onSelectionChanged: (value) =>
                  setState(() => register = value.first)),
          if (register) ...[
            const SizedBox(height: 14),
            SegmentedButton<String>(
                segments: const [
                  ButtonSegment(value: 'buyer', label: Text('采购商')),
                  ButtonSegment(value: 'vendor', label: Text('厂商'))
                ],
                selected: {
                  role
                },
                onSelectionChanged: (value) =>
                    setState(() => role = value.first))
          ],
          const SizedBox(height: 16),
          TextField(
              controller: username,
              autofillHints: const [AutofillHints.username],
              decoration: const InputDecoration(labelText: '账号')),
          const SizedBox(height: 12),
          if (register && role == 'vendor') ...[
            TextField(
                controller: company,
                decoration: const InputDecoration(labelText: '公司全称')),
            const SizedBox(height: 12)
          ],
          if (register && role == 'buyer') ...[
            TextField(
                controller: displayName,
                decoration: const InputDecoration(labelText: '称呼')),
            const SizedBox(height: 12),
            TextField(
                controller: contact,
                decoration: const InputDecoration(labelText: '联系人')),
            const SizedBox(height: 12),
            TextField(
                controller: phone,
                keyboardType: TextInputType.phone,
                decoration: const InputDecoration(labelText: '联系电话')),
            const SizedBox(height: 12)
          ],
          TextField(
              controller: password,
              obscureText: true,
              autofillHints: const [AutofillHints.password],
              decoration: const InputDecoration(labelText: '密码')),
          if (register) ...[
            const SizedBox(height: 12),
            TextField(
                controller: confirm,
                obscureText: true,
                decoration: const InputDecoration(labelText: '确认密码'))
          ],
          const SizedBox(height: 18),
          if (auth.hasError)
            Text(auth.error.toString(),
                style: TextStyle(color: Theme.of(context).colorScheme.error)),
          FilledButton(
              onPressed: auth.isLoading ? null : submit,
              child: Text(auth.isLoading
                  ? '正在提交…'
                  : register
                      ? '创建账号'
                      : '登录'))
        ]));
  }
}

final ownPostsProvider = FutureProvider<PageData<MarketPost>>((ref) async {
  final data = await ref
      .watch(apiClientProvider)
      .getJson('/api/v1/me/market-posts', query: {'pageSize': 40});
  return ApiClient.page(data, MarketPost.fromJson);
});

class MyPage extends ConsumerWidget {
  const MyPage({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final account = ref.watch(authControllerProvider).valueOrNull;
    if (account == null) {
      return Scaffold(
          appBar: AppBar(title: const Text('我的')),
          body: Center(
              child: FilledButton.icon(
                  onPressed: () => context.push('/login?returnTo=/me'),
                  icon: const Icon(Icons.login),
                  label: const Text('登录或注册'))));
    }
    return Scaffold(
      appBar: AppBar(title: const Text('我的'), actions: [
        IconButton(
            onPressed: () => ref.read(authControllerProvider.notifier).logout(),
            icon: const Icon(Icons.logout))
      ]),
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(ownPostsProvider.future),
        child: ListView(padding: const EdgeInsets.all(14), children: [
          Card(
              child: ListTile(
                  leading: CircleAvatar(
                      child: Text(account.username.isEmpty
                          ? '农'
                          : account.username.substring(0, 1))),
                  title: Text(account.username),
                  subtitle: Text(account.isVendor ? '厂商账号' : '采购商账号'))),
          const SizedBox(height: 12),
          if (account.isVendor) ...[
            Card(
                child: ListTile(
                    onTap: () => context.push('/vendor-center'),
                    leading: const Icon(Icons.factory_outlined),
                    title: const Text('厂家资料'),
                    subtitle: const Text('维护企业资料和提审状态'),
                    trailing: const Icon(Icons.chevron_right))),
            Card(
                child: ListTile(
                    onTap: () => context.push('/vendor-products'),
                    leading: const Icon(Icons.inventory_2_outlined),
                    title: const Text('产品管理'),
                    subtitle: const Text('维护自身产品和供应信息'),
                    trailing: const Icon(Icons.chevron_right))),
          ] else ...[
            Card(
                child: ListTile(
                    onTap: () => context.push('/buyer-profile'),
                    leading: const Icon(Icons.badge_outlined),
                    title: const Text('采购商资料'),
                    subtitle: const Text('维护联系人、电话和地区'),
                    trailing: const Icon(Icons.chevron_right))),
          ],
          Card(
              child: ListTile(
                  onTap: () => context.go('/publish'),
                  leading: const Icon(Icons.add_circle_outline),
                  title: Text(account.isVendor ? '发布供应' : '发布求购'),
                  trailing: const Icon(Icons.chevron_right))),
          Padding(
              padding: const EdgeInsets.fromLTRB(4, 22, 4, 10),
              child: Text('我的发布',
                  style: Theme.of(context)
                      .textTheme
                      .titleLarge
                      ?.copyWith(fontWeight: FontWeight.bold))),
          ref.watch(ownPostsProvider).when(
                loading: () => const Center(child: CircularProgressIndicator()),
                error: (error, _) => Text('加载失败：$error'),
                data: (page) => page.items.isEmpty
                    ? const Padding(
                        padding: EdgeInsets.all(30),
                        child: Center(child: Text('还没有发布信息')))
                    : Column(
                        children: page.items
                            .map((post) => Padding(
                                padding: const EdgeInsets.only(bottom: 10),
                                child: MarketPostCard(post,
                                    trailing: PopupMenuButton<String>(
                                        onSelected: (value) async {
                                          if (value == 'edit') {
                                            context.push('/publish/${post.id}');
                                          }
                                          if (value == 'withdraw') {
                                            await ref
                                                .read(apiClientProvider)
                                                .deleteJson(
                                                    '/api/v1/me/market-posts/${post.id}');
                                            ref.invalidate(ownPostsProvider);
                                          }
                                        },
                                        itemBuilder: (_) => const [
                                              PopupMenuItem(
                                                  value: 'edit',
                                                  child: Text('编辑')),
                                              PopupMenuItem(
                                                  value: 'withdraw',
                                                  child: Text('撤回'))
                                            ]))))
                            .toList()),
              ),
        ]),
      ),
    );
  }
}

class BuyerProfilePage extends ConsumerStatefulWidget {
  const BuyerProfilePage({super.key});

  @override
  ConsumerState<BuyerProfilePage> createState() => _BuyerProfilePageState();
}

class _BuyerProfilePageState extends ConsumerState<BuyerProfilePage> {
  final displayName = TextEditingController();
  final contactName = TextEditingController();
  final phone = TextEditingController();
  final province = TextEditingController();
  final city = TextEditingController();
  bool loading = true;
  String? message;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    displayName.dispose();
    contactName.dispose();
    phone.dispose();
    province.dispose();
    city.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final data =
          await ref.read(apiClientProvider).getJson('/api/v1/me/profile');
      final profile = data['profile'] as Map<String, dynamic>? ?? const {};
      displayName.text = profile['displayName'] as String? ?? '';
      contactName.text = profile['contactName'] as String? ?? '';
      phone.text = profile['phone'] as String? ?? '';
      province.text = profile['province'] as String? ?? '';
      city.text = profile['city'] as String? ?? '';
    } catch (error) {
      message = error.toString();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> _save() async {
    setState(() {
      loading = true;
      message = null;
    });
    try {
      await ref.read(apiClientProvider).putJson('/api/v1/me/profile', {
        'displayName': displayName.text,
        'contactName': contactName.text,
        'phone': phone.text,
        'province': province.text,
        'city': city.text,
      });
      message = '采购商资料已保存';
    } catch (error) {
      message = error.toString();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
      appBar: AppBar(title: const Text('采购商资料')),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : ListView(padding: const EdgeInsets.all(14), children: [
              if (message != null) Text(message!),
              TextField(
                  controller: displayName,
                  decoration: const InputDecoration(labelText: '称呼')),
              const SizedBox(height: 12),
              TextField(
                  controller: contactName,
                  decoration: const InputDecoration(labelText: '联系人')),
              const SizedBox(height: 12),
              TextField(
                  controller: phone,
                  keyboardType: TextInputType.phone,
                  decoration: const InputDecoration(labelText: '联系电话')),
              const SizedBox(height: 12),
              TextField(
                  controller: province,
                  decoration: const InputDecoration(labelText: '省份')),
              const SizedBox(height: 12),
              TextField(
                  controller: city,
                  decoration: const InputDecoration(labelText: '城市')),
              const SizedBox(height: 18),
              FilledButton(onPressed: _save, child: const Text('保存资料'))
            ]));
}

class VendorCenterPage extends ConsumerStatefulWidget {
  const VendorCenterPage({super.key});
  @override
  ConsumerState<VendorCenterPage> createState() => _VendorCenterPageState();
}

class _VendorCenterPageState extends ConsumerState<VendorCenterPage> {
  final name = TextEditingController();
  final mainProducts = TextEditingController();
  final address = TextEditingController();
  final description = TextEditingController();
  bool loading = true;
  String? message;
  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final data = await ref
          .read(apiClientProvider)
          .getJson('/api/admin/vendor-profile');
      final vendor = (data['draft'] ?? data['vendor']) as Map<String, dynamic>;
      name.text = vendor['name'] as String? ?? '';
      mainProducts.text = vendor['mainProducts'] as String? ?? '';
      address.text = vendor['address'] as String? ?? '';
      description.text = vendor['description'] as String? ?? '';
    } catch (error) {
      message = error.toString();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> _save() async {
    setState(() => loading = true);
    try {
      await ref.read(apiClientProvider).putJson('/api/admin/vendor-profile', {
        'name': name.text,
        'mainProducts': mainProducts.text,
        'address': address.text,
        'description': description.text
      });
      message = '资料已提交审核';
    } catch (error) {
      message = error.toString();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
      appBar: AppBar(title: const Text('厂家资料')),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : ListView(padding: const EdgeInsets.all(14), children: [
              if (message != null) Text(message!),
              TextField(
                  controller: name,
                  decoration: const InputDecoration(labelText: '厂家名称')),
              const SizedBox(height: 12),
              TextField(
                  controller: mainProducts,
                  decoration: const InputDecoration(labelText: '主营产品')),
              const SizedBox(height: 12),
              TextField(
                  controller: address,
                  decoration: const InputDecoration(labelText: '详细地址')),
              const SizedBox(height: 12),
              TextField(
                  controller: description,
                  maxLines: 6,
                  decoration: const InputDecoration(labelText: '厂家介绍')),
              const SizedBox(height: 18),
              FilledButton(onPressed: _save, child: const Text('提交审核'))
            ]));
}

class VendorProductsPage extends ConsumerStatefulWidget {
  const VendorProductsPage({super.key});

  @override
  ConsumerState<VendorProductsPage> createState() => _VendorProductsPageState();
}

class _VendorProductsPageState extends ConsumerState<VendorProductsPage> {
  late Future<List<dynamic>> products;

  @override
  void initState() {
    super.initState();
    products = _load();
  }

  Future<List<dynamic>> _load() =>
      ref.read(apiClientProvider).getList('/api/admin/vendor-products');

  void _reload() => setState(() => products = _load());

  Future<void> _edit({Map<String, dynamic>? row}) async {
    final product = row?['product'] as Map<String, dynamic>? ?? const {};
    final supplier = row?['supplier'] as Map<String, dynamic>? ?? const {};
    final name = TextEditingController(
        text: (supplier['vendorProductName'] ?? product['name']) as String? ??
            '');
    final model = TextEditingController(
        text: (supplier['vendorModel'] ?? product['compatibleModels'])
                as String? ??
            '');
    final price =
        TextEditingController(text: supplier['priceNote'] as String? ?? '');
    final description =
        TextEditingController(text: supplier['description'] as String? ?? '');
    String? error;
    var saving = false;

    await showDialog<void>(
      context: context,
      builder: (dialogContext) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: Text(row == null ? '提交新产品' : '编辑供应信息'),
          content: SingleChildScrollView(
            child: Column(mainAxisSize: MainAxisSize.min, children: [
              TextField(
                  controller: name,
                  decoration: const InputDecoration(labelText: '产品名称')),
              const SizedBox(height: 10),
              TextField(
                  controller: model,
                  decoration: const InputDecoration(labelText: '型号 / 适配机型')),
              const SizedBox(height: 10),
              TextField(
                  controller: price,
                  decoration: const InputDecoration(labelText: '价格说明')),
              const SizedBox(height: 10),
              TextField(
                  controller: description,
                  maxLines: 4,
                  decoration: const InputDecoration(labelText: '供应说明')),
              if (error != null) ...[
                const SizedBox(height: 10),
                Text(error!, style: const TextStyle(color: Colors.red))
              ]
            ]),
          ),
          actions: [
            TextButton(
                onPressed: saving ? null : () => Navigator.pop(dialogContext),
                child: const Text('取消')),
            FilledButton(
              onPressed: saving
                  ? null
                  : () async {
                      if (name.text.trim().isEmpty) {
                        setDialogState(() => error = '请填写产品名称');
                        return;
                      }
                      setDialogState(() {
                        saving = true;
                        error = null;
                      });
                      try {
                        final body = <String, dynamic>{
                          if (row == null) 'name': name.text.trim(),
                          if (row == null)
                            'compatibleModels': model.text.trim(),
                          if (row != null)
                            'vendorProductName': name.text.trim(),
                          if (row != null) 'vendorModel': model.text.trim(),
                          'description': description.text.trim(),
                          'priceNote': price.text.trim(),
                        };
                        if (row == null) {
                          await ref
                              .read(apiClientProvider)
                              .postJson('/api/admin/vendor-products', body);
                        } else {
                          await ref.read(apiClientProvider).putJson(
                              '/api/admin/vendor-products/${supplier['id']}',
                              body);
                        }
                        if (dialogContext.mounted) Navigator.pop(dialogContext);
                        _reload();
                      } catch (exception) {
                        setDialogState(() {
                          saving = false;
                          error = exception.toString();
                        });
                      }
                    },
              child: Text(saving ? '提交中…' : '提交审核'),
            )
          ],
        ),
      ),
    );
    name.dispose();
    model.dispose();
    price.dispose();
    description.dispose();
  }

  Future<void> _disable(Map<String, dynamic> supplier) async {
    await ref
        .read(apiClientProvider)
        .deleteJson('/api/admin/vendor-products/${supplier['id']}');
    _reload();
  }

  @override
  Widget build(BuildContext context) => Scaffold(
      appBar: AppBar(title: const Text('我的产品'), actions: [
        IconButton(
            tooltip: '提交新产品',
            onPressed: () => _edit(),
            icon: const Icon(Icons.add_circle_outline))
      ]),
      body: FutureBuilder<List<dynamic>>(
          future: products,
          builder: (context, snapshot) {
            if (!snapshot.hasData) {
              return snapshot.hasError
                  ? Center(child: Text('加载失败：${snapshot.error}'))
                  : const Center(child: CircularProgressIndicator());
            }
            final items = snapshot.data!;
            return items.isEmpty
                ? Center(
                    child: FilledButton.icon(
                        onPressed: () => _edit(),
                        icon: const Icon(Icons.add),
                        label: const Text('提交第一个产品')))
                : ListView.separated(
                    padding: const EdgeInsets.all(14),
                    itemCount: items.length,
                    separatorBuilder: (_, __) => const SizedBox(height: 9),
                    itemBuilder: (_, index) {
                      final row = items[index] as Map<String, dynamic>;
                      final product =
                          row['product'] as Map<String, dynamic>? ?? const {};
                      final supplier =
                          row['supplier'] as Map<String, dynamic>? ?? const {};
                      return Card(
                          child: ListTile(
                              title:
                                  Text(product['name'] as String? ?? '未命名产品'),
                              subtitle: Text(
                                  '状态：${supplier['status'] ?? 'pending'}\n${supplier['priceNote'] ?? ''}'),
                              isThreeLine: true,
                              trailing: PopupMenuButton<String>(
                                  onSelected: (action) {
                                    if (action == 'edit') _edit(row: row);
                                    if (action == 'disable') {
                                      _disable(supplier);
                                    }
                                  },
                                  itemBuilder: (_) => const [
                                        PopupMenuItem(
                                            value: 'edit', child: Text('编辑')),
                                        PopupMenuItem(
                                            value: 'disable',
                                            child: Text('停止供应'))
                                      ])));
                    });
          }));
}
