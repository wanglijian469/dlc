import 'dart:convert';
import 'dart:io';

import 'package:dlc_mobile/core/auth/session_store.dart';
import 'package:dlc_mobile/core/network/api_client.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';

class _MemorySessionStore extends SessionStore {
  _MemorySessionStore(this.tokens) : super(const FlutterSecureStorage());

  SessionTokens? tokens;

  @override
  Future<SessionTokens?> read() async => tokens;

  @override
  Future<void> write(SessionTokens next) async => tokens = next;

  @override
  Future<void> clear() async => tokens = null;
}

void main() {
  late HttpServer server;
  late Uri origin;

  setUp(() async {
    server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
    origin = Uri.parse('http://${server.address.address}:${server.port}');
  });

  tearDown(() => server.close(force: true));

  test('rotates refresh token and replays one unauthorized request', () async {
    final store = _MemorySessionStore(const SessionTokens(
        accessToken: 'expired-access', refreshToken: 'old-refresh'));
    var refreshCalls = 0;
    server.listen((request) async {
      if (request.uri.path == '/api/v1/app/auth/refresh') {
        refreshCalls++;
        final payload = jsonDecode(await utf8.decoder.bind(request).join())
            as Map<String, dynamic>;
        expect(payload['refreshToken'], 'old-refresh');
        _reply(request.response, {
          'code': 0,
          'message': 'ok',
          'data': {'accessToken': 'new-access', 'refreshToken': 'new-refresh'}
        });
        return;
      }
      if (request.headers.value(HttpHeaders.authorizationHeader) !=
          'Bearer new-access') {
        request.response.statusCode = HttpStatus.unauthorized;
        _reply(request.response,
            {'code': 401, 'message': 'expired', 'data': <String, dynamic>{}});
        return;
      }
      _reply(request.response, {
        'code': 0,
        'message': 'ok',
        'data': {'id': 7}
      });
    });

    final result = await ApiClient(store, baseUrl: origin.toString())
        .getJson('/api/v1/app/auth/me');

    expect(result['id'], 7);
    expect(refreshCalls, 1);
    expect(store.tokens?.accessToken, 'new-access');
    expect(store.tokens?.refreshToken, 'new-refresh');
  });

  test('retries a safe GET once after a transient gateway failure', () async {
    final store = _MemorySessionStore(null);
    var attempts = 0;
    server.listen((request) {
      attempts++;
      if (attempts == 1) {
        request.response.statusCode = HttpStatus.serviceUnavailable;
        _reply(request.response, {
          'code': 503,
          'message': 'weak network',
          'data': <String, dynamic>{}
        });
        return;
      }
      _reply(request.response, {
        'code': 0,
        'message': 'ok',
        'data': {'items': <dynamic>[]}
      });
    });

    final result = await ApiClient(store, baseUrl: origin.toString())
        .getJson('/api/products');

    expect(result['items'], isEmpty);
    expect(attempts, 2);
  });
}

void _reply(HttpResponse response, Map<String, dynamic> body) {
  response.headers.contentType = ContentType.json;
  response.write(jsonEncode(body));
  response.close();
}
