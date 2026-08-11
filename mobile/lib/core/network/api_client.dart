import 'dart:async';

import 'package:dio/dio.dart';

import '../auth/session_store.dart';
import '../config/app_config.dart';
import '../models/models.dart';

class ApiException implements Exception {
  const ApiException(this.message, {this.statusCode});
  final String message;
  final int? statusCode;
  @override
  String toString() => message;
}

class ApiClient {
  ApiClient(this._sessionStore, {String? baseUrl})
      : _baseUrl = baseUrl ?? AppConfig.apiBaseUrl,
        _dio = Dio(BaseOptions(
            baseUrl: baseUrl ?? AppConfig.apiBaseUrl,
            connectTimeout: const Duration(seconds: 12),
            receiveTimeout: const Duration(seconds: 20))) {
    _dio.interceptors.add(
        QueuedInterceptorsWrapper(onRequest: _authorize, onError: _recover));
  }

  final SessionStore _sessionStore;
  final String _baseUrl;
  final Dio _dio;
  Future<SessionTokens?>? _refreshing;

  Future<void> _authorize(
      RequestOptions options, RequestInterceptorHandler handler) async {
    final session = await _sessionStore.read();
    if (session != null) {
      options.headers['Authorization'] = 'Bearer ${session.accessToken}';
    }
    handler.next(options);
  }

  Future<void> _recover(
      DioException error, ErrorInterceptorHandler handler) async {
    final request = error.requestOptions;
    final transientFailure = error.type == DioExceptionType.connectionTimeout ||
        error.type == DioExceptionType.receiveTimeout ||
        error.type == DioExceptionType.connectionError ||
        const {502, 503, 504}.contains(error.response?.statusCode);
    if (request.method == 'GET' &&
        transientFailure &&
        request.extra['networkRetried'] != true) {
      request.extra['networkRetried'] = true;
      await Future<void>.delayed(const Duration(milliseconds: 400));
      try {
        handler.resolve(await _dio.fetch<dynamic>(request));
        return;
      } on DioException {
        // The original failure is returned after the single safe retry.
      }
    }
    if (error.response?.statusCode != 401 ||
        request.extra['retried'] == true ||
        _isCredentialEndpoint(request.path)) {
      handler.next(error);
      return;
    }
    try {
      final tokens = await (_refreshing ??= _refresh());
      _refreshing = null;
      if (tokens == null) throw const ApiException('登录已过期');
      request.extra['retried'] = true;
      request.headers['Authorization'] = 'Bearer ${tokens.accessToken}';
      handler.resolve(await _dio.fetch<dynamic>(request));
    } catch (_) {
      _refreshing = null;
      await _sessionStore.clear();
      handler.next(error);
    }
  }

  bool _isCredentialEndpoint(String path) =>
      path.endsWith('/login') ||
      path.endsWith('/register') ||
      path.endsWith('/refresh');

  Future<SessionTokens?> _refresh() async {
    final session = await _sessionStore.read();
    if (session == null) return null;
    final plain = Dio(BaseOptions(baseUrl: _baseUrl));
    final response = await plain.post<Map<String, dynamic>>(
        '/api/v1/app/auth/refresh',
        data: {'refreshToken': session.refreshToken});
    final data = _unwrap(response.data)['data'] as Map<String, dynamic>;
    final tokens = SessionTokens(
        accessToken: data['accessToken'] as String,
        refreshToken: data['refreshToken'] as String);
    await _sessionStore.write(tokens);
    return tokens;
  }

  Future<Map<String, dynamic>> getJson(String path,
          {Map<String, dynamic>? query}) async =>
      _data(await _dio.get<Map<String, dynamic>>(path, queryParameters: query));
  Future<List<dynamic>> getList(String path,
      {Map<String, dynamic>? query}) async {
    final response =
        await _dio.get<Map<String, dynamic>>(path, queryParameters: query);
    final value = _unwrap(response.data)['data'];
    if (value is! List) throw const ApiException('服务器列表响应格式不正确');
    return value;
  }

  Future<Map<String, dynamic>> postJson(String path, Object? body) async =>
      _data(await _dio.post<Map<String, dynamic>>(path, data: body));
  Future<Map<String, dynamic>> putJson(String path, Object? body) async =>
      _data(await _dio.put<Map<String, dynamic>>(path, data: body));
  Future<Map<String, dynamic>> deleteJson(String path) async =>
      _data(await _dio.delete<Map<String, dynamic>>(path));

  Future<int> uploadImage(String path) async {
    final name = path.split(RegExp(r'[/\\]')).last;
    final response = await _dio.post<Map<String, dynamic>>('/api/v1/media',
        data: FormData.fromMap(
            {'file': await MultipartFile.fromFile(path, filename: name)}));
    return (_data(response)['assetId'] as num).toInt();
  }

  Map<String, dynamic> _data(Response<Map<String, dynamic>> response) =>
      _unwrap(response.data)['data'] as Map<String, dynamic>;

  static Map<String, dynamic> _unwrap(Map<String, dynamic>? body) {
    if (body == null) throw const ApiException('服务器响应为空');
    if (body['code'] != 0) {
      throw ApiException(body['message'] as String? ?? '请求失败');
    }
    return body;
  }

  static PageData<T> page<T>(Map<String, dynamic> data,
          T Function(Map<String, dynamic>) convert) =>
      PageData<T>(
        items: (data['items'] as List? ?? const [])
            .map((item) => convert(item as Map<String, dynamic>))
            .toList(),
        page: (data['page'] as num?)?.toInt() ?? 1,
        pageSize: (data['pageSize'] as num?)?.toInt() ?? 20,
        total: (data['total'] as num?)?.toInt() ?? 0,
      );
}
