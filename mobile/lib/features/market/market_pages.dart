import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';

import '../../core/auth/auth_controller.dart';
import '../../core/models/models.dart';
import '../../core/network/api_client.dart';
import '../../core/widgets/remote_image.dart';

final marketPostsProvider =
    FutureProvider.family<PageData<MarketPost>, String?>((ref, type) async {
  final data = await ref
      .watch(apiClientProvider)
      .getJson('/api/v1/market-posts', query: {'type': type, 'pageSize': 40});
  return ApiClient.page(data, MarketPost.fromJson);
});

class MarketPostsPage extends ConsumerStatefulWidget {
  const MarketPostsPage({super.key});
  @override
  ConsumerState<MarketPostsPage> createState() => _MarketPostsPageState();
}

class _MarketPostsPageState extends ConsumerState<MarketPostsPage> {
  String? type;
  @override
  Widget build(BuildContext context) => Scaffold(
      appBar: AppBar(title: const Text('供求信息')),
      body: Column(children: [
        Padding(
            padding: const EdgeInsets.all(12),
            child: Row(children: [
              ChoiceChip(
                  label: const Text('全部'),
                  selected: type == null,
                  onSelected: (_) => setState(() => type = null)),
              const SizedBox(width: 8),
              ChoiceChip(
                  label: const Text('求购'),
                  selected: type == 'demand',
                  onSelected: (_) => setState(() => type = 'demand')),
              const SizedBox(width: 8),
              ChoiceChip(
                  label: const Text('供应'),
                  selected: type == 'supply',
                  onSelected: (_) => setState(() => type = 'supply'))
            ])),
        Expanded(
            child: ref.watch(marketPostsProvider(type)).when(
                loading: () => const Center(child: CircularProgressIndicator()),
                error: (error, _) => Center(
                    child: TextButton(
                        onPressed: () =>
                            ref.invalidate(marketPostsProvider(type)),
                        child: Text('加载失败：$error'))),
                data: (page) => RefreshIndicator(
                    onRefresh: () =>
                        ref.refresh(marketPostsProvider(type).future),
                    child: page.items.isEmpty
                        ? ListView(children: const [
                            SizedBox(height: 180),
                            Center(child: Text('暂无供求信息'))
                          ])
                        : ListView.separated(
                            padding: const EdgeInsets.all(12),
                            itemCount: page.items.length,
                            separatorBuilder: (_, __) =>
                                const SizedBox(height: 10),
                            itemBuilder: (_, index) =>
                                MarketPostCard(page.items[index])))))
      ]));
}

class MarketPostCard extends StatelessWidget {
  const MarketPostCard(this.post, {this.trailing, super.key});
  final MarketPost post;
  final Widget? trailing;
  @override
  Widget build(BuildContext context) => Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
          onTap: () => context.push('/market/${post.id}'),
          child: Row(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
            SizedBox(
                width: 112,
                height: 136,
                child: RemoteImage(
                    post.images.isEmpty ? null : post.images.first)),
            Expanded(
                child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(children: [
                            Container(
                                decoration: BoxDecoration(
                                    color: post.isDemand
                                        ? const Color(0xFFFFEAD0)
                                        : const Color(0xFFDFF5E9),
                                    borderRadius: BorderRadius.circular(20)),
                                padding: const EdgeInsets.symmetric(
                                    horizontal: 8, vertical: 4),
                                child: Text(post.isDemand ? '求购' : '供应',
                                    style: const TextStyle(
                                        fontSize: 11,
                                        fontWeight: FontWeight.bold))),
                            const Spacer(),
                            if (trailing != null) trailing!
                          ]),
                          const SizedBox(height: 7),
                          Text(post.title,
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                              style:
                                  const TextStyle(fontWeight: FontWeight.bold)),
                          const Spacer(),
                          Text(
                              [post.province, post.city]
                                      .whereType<String>()
                                      .join(' · ')
                                      .isEmpty
                                  ? '全国'
                                  : [post.province, post.city]
                                      .whereType<String>()
                                      .join(' · '),
                              style: Theme.of(context).textTheme.labelSmall),
                          const SizedBox(height: 5),
                          Text(post.quantity ?? '数量面议',
                              style: TextStyle(
                                  color:
                                      Theme.of(context).colorScheme.secondary,
                                  fontWeight: FontWeight.bold))
                        ])))
          ])));
}

class MarketPostDetailPage extends ConsumerStatefulWidget {
  const MarketPostDetailPage(this.id, {super.key});
  final String id;
  @override
  ConsumerState<MarketPostDetailPage> createState() =>
      _MarketPostDetailPageState();
}

class _MarketPostDetailPageState extends ConsumerState<MarketPostDetailPage> {
  Map<String, dynamic>? contact;
  String? error;

  Future<void> _showContact() async {
    if (ref.read(authControllerProvider).valueOrNull == null) {
      context.push('/login?returnTo=/market/${widget.id}');
      return;
    }
    try {
      final value = await ref
          .read(apiClientProvider)
          .getJson('/api/v1/market-posts/${widget.id}/contact');
      if (mounted) setState(() => contact = value);
    } catch (reason) {
      if (mounted) setState(() => error = reason.toString());
    }
  }

  @override
  Widget build(BuildContext context) {
    final api = ref.read(apiClientProvider);
    return Scaffold(
        appBar: AppBar(title: const Text('供求详情')),
        bottomNavigationBar: SafeArea(
            minimum: const EdgeInsets.fromLTRB(14, 8, 14, 12),
            child: FilledButton.icon(
                onPressed: _showContact,
                icon: const Icon(Icons.phone),
                label: Text(contact == null ? '登录后查看电话' : '联系方式已显示'))),
        body: FutureBuilder<Map<String, dynamic>>(
            future: api.getJson('/api/v1/market-posts/${widget.id}'),
            builder: (context, snapshot) {
              if (!snapshot.hasData) {
                return snapshot.hasError
                    ? Center(child: Text('加载失败：${snapshot.error}'))
                    : const Center(child: CircularProgressIndicator());
              }
              final post = MarketPost.fromJson(snapshot.data!);
              return ListView(
                  padding: const EdgeInsets.fromLTRB(14, 14, 14, 90),
                  children: [
                    if (post.images.isNotEmpty)
                      SizedBox(
                          height: 240,
                          child: PageView(
                              children: post.images
                                  .map((image) => ClipRRect(
                                      borderRadius: BorderRadius.circular(14),
                                      child: RemoteImage(image,
                                          width: double.infinity)))
                                  .toList())),
                    const SizedBox(height: 16),
                    Row(children: [
                      Chip(label: Text(post.isDemand ? '求购信息' : '供应信息')),
                      const Spacer(),
                      Text(post.province ?? '全国')
                    ]),
                    Text(post.title,
                        style: Theme.of(context)
                            .textTheme
                            .headlineSmall
                            ?.copyWith(fontWeight: FontWeight.bold)),
                    const SizedBox(height: 18),
                    _InfoRow('数量', post.quantity ?? '面议'),
                    _InfoRow('适配机型', post.compatibleModels ?? '请联系发布者'),
                    _InfoRow('交期', post.deliveryNote ?? '双方协商'),
                    const Divider(height: 32),
                    Text('详细说明',
                        style: Theme.of(context)
                            .textTheme
                            .titleMedium
                            ?.copyWith(fontWeight: FontWeight.bold)),
                    const SizedBox(height: 8),
                    Text(post.description, style: const TextStyle(height: 1.7)),
                    const SizedBox(height: 18),
                    Text('发布者：${post.publisherName}'),
                    const SizedBox(height: 22),
                    Card(
                        color: Theme.of(context).colorScheme.primaryContainer,
                        child: Padding(
                            padding: const EdgeInsets.all(16),
                            child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  const Text('联系方式受平台保护',
                                      style: TextStyle(
                                          fontWeight: FontWeight.bold)),
                                  const SizedBox(height: 7),
                                  const Text('登录后查看，异常频率会被限制。'),
                                  if (contact != null) ...[
                                    const SizedBox(height: 12),
                                    Text(
                                        '${contact!['contactName']}  ${contact!['phone']}',
                                        style: const TextStyle(
                                            fontWeight: FontWeight.bold))
                                  ],
                                  if (error != null)
                                    Text(error!,
                                        style: TextStyle(
                                            color: Theme.of(context)
                                                .colorScheme
                                                .error))
                                ])))
                  ]);
            }));
  }
}

class _InfoRow extends StatelessWidget {
  const _InfoRow(this.label, this.value);
  final String label, value;
  @override
  Widget build(BuildContext context) => Padding(
      padding: const EdgeInsets.symmetric(vertical: 7),
      child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
        SizedBox(
            width: 85,
            child: Text(label, style: Theme.of(context).textTheme.bodySmall)),
        Expanded(
            child: Text(value,
                style: const TextStyle(fontWeight: FontWeight.w600)))
      ]));
}

class PublishPage extends ConsumerStatefulWidget {
  const PublishPage({this.id, super.key});
  final String? id;
  @override
  ConsumerState<PublishPage> createState() => _PublishPageState();
}

class _PublishPageState extends ConsumerState<PublishPage> {
  final formKey = GlobalKey<FormState>();
  final title = TextEditingController();
  final models = TextEditingController();
  final province = TextEditingController();
  final city = TextEditingController();
  final quantity = TextEditingController();
  final delivery = TextEditingController();
  final description = TextEditingController();
  final contact = TextEditingController();
  final phone = TextEditingController();
  final imagePicker = ImagePicker();
  final assetIds = <int>[];
  final imagePaths = <String>[];
  bool saving = false;
  String? error;
  bool loaded = false;
  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (!loaded && widget.id != null) {
      loaded = true;
      _loadExisting();
    }
  }

  Future<void> _loadExisting() async {
    try {
      final data = await ref
          .read(apiClientProvider)
          .getJson('/api/v1/me/market-posts/${widget.id}');
      final post = MarketPost.fromJson(data['post'] as Map<String, dynamic>);
      title.text = post.title;
      models.text = post.compatibleModels ?? '';
      province.text = post.province ?? '';
      city.text = post.city ?? '';
      quantity.text = post.quantity ?? '';
      delivery.text = post.deliveryNote ?? '';
      description.text = post.description;
      contact.text = data['contactName'] as String? ?? '';
      phone.text = data['contactPhone'] as String? ?? '';
      assetIds.addAll((data['assetIds'] as List? ?? const [])
          .map((value) => (value as num).toInt()));
      setState(() {});
    } catch (reason) {
      setState(() => error = reason.toString());
    }
  }

  Future<void> _pick() async {
    final files = await imagePicker.pickMultiImage(imageQuality: 86);
    if (assetIds.length + files.length > 6) {
      setState(() => error = '最多上传 6 张图片');
      return;
    }
    for (final file in files) {
      final id = await ref.read(apiClientProvider).uploadImage(file.path);
      assetIds.add(id);
      imagePaths.add(file.path);
    }
    setState(() {});
  }

  Future<void> _submit() async {
    if (!formKey.currentState!.validate()) return;
    final account = ref.read(authControllerProvider).valueOrNull;
    if (account == null) {
      context.push('/login?returnTo=/publish');
      return;
    }
    setState(() {
      saving = true;
      error = null;
    });
    final body = {
      'type': account.isVendor ? 'supply' : 'demand',
      'title': title.text,
      'compatibleModels': models.text,
      'province': province.text,
      'city': city.text,
      'quantity': quantity.text,
      'deliveryNote': delivery.text,
      'description': description.text,
      'contactName': contact.text,
      'contactPhone': phone.text,
      'expiresInDays': 30,
      'assetIds': assetIds
    };
    try {
      final data = widget.id == null
          ? await ref
              .read(apiClientProvider)
              .postJson('/api/v1/me/market-posts', body)
          : await ref
              .read(apiClientProvider)
              .putJson('/api/v1/me/market-posts/${widget.id}', body);
      if (mounted) context.go('/market/${data['id']}');
    } catch (reason) {
      setState(() => error = reason.toString());
    } finally {
      if (mounted) setState(() => saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final account = ref.watch(authControllerProvider).valueOrNull;
    if (account == null) {
      return Scaffold(
          appBar: AppBar(title: const Text('发布')),
          body: Center(
              child: FilledButton(
                  onPressed: () => context.push('/login?returnTo=/publish'),
                  child: const Text('登录后发布'))));
    }
    return Scaffold(
        appBar: AppBar(
            title: Text(widget.id == null
                ? (account.isVendor ? '发布供应' : '发布求购')
                : '编辑供求')),
        body: Form(
            key: formKey,
            child: ListView(padding: const EdgeInsets.all(14), children: [
              if (error != null)
                Text(error!,
                    style:
                        TextStyle(color: Theme.of(context).colorScheme.error)),
              TextFormField(
                  controller: title,
                  decoration: const InputDecoration(labelText: '标题'),
                  validator: (value) =>
                      (value ?? '').trim().length < 4 ? '标题至少 4 个字' : null),
              const SizedBox(height: 12),
              Row(children: [
                Expanded(
                    child: TextFormField(
                        controller: province,
                        decoration: const InputDecoration(labelText: '省份'))),
                const SizedBox(width: 10),
                Expanded(
                    child: TextFormField(
                        controller: city,
                        decoration: const InputDecoration(labelText: '城市')))
              ]),
              const SizedBox(height: 12),
              TextFormField(
                  controller: models,
                  decoration: const InputDecoration(labelText: '适配机型')),
              const SizedBox(height: 12),
              TextFormField(
                  controller: quantity,
                  decoration: const InputDecoration(labelText: '数量')),
              const SizedBox(height: 12),
              TextFormField(
                  controller: delivery,
                  decoration: const InputDecoration(labelText: '交期说明')),
              const SizedBox(height: 12),
              TextFormField(
                  controller: description,
                  maxLines: 6,
                  decoration: const InputDecoration(labelText: '详细说明'),
                  validator: (value) =>
                      (value ?? '').trim().length < 10 ? '详细说明至少 10 个字' : null),
              const SizedBox(height: 12),
              TextFormField(
                  controller: contact,
                  decoration: const InputDecoration(labelText: '联系人'),
                  validator: (value) =>
                      (value ?? '').trim().isEmpty ? '请填写联系人' : null),
              const SizedBox(height: 12),
              TextFormField(
                  controller: phone,
                  keyboardType: TextInputType.phone,
                  decoration: const InputDecoration(labelText: '联系电话'),
                  validator: (value) =>
                      (value ?? '').trim().length < 6 ? '请填写有效电话' : null),
              const SizedBox(height: 14),
              OutlinedButton.icon(
                  onPressed: _pick,
                  icon: const Icon(Icons.add_photo_alternate_outlined),
                  label: Text('添加图片（${assetIds.length}/6）')),
              const SizedBox(height: 20),
              FilledButton.icon(
                  onPressed: saving ? null : _submit,
                  icon: const Icon(Icons.send),
                  label: Text(saving ? '正在发布…' : '立即发布'))
            ])));
  }
}
