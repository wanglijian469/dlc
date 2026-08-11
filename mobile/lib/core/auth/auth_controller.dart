import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../models/models.dart';
import '../network/api_client.dart';
import 'session_store.dart';

const _secureStorage = FlutterSecureStorage(
    aOptions: AndroidOptions(encryptedSharedPreferences: true));
final sessionStoreProvider =
    Provider<SessionStore>((ref) => const SessionStore(_secureStorage));
final apiClientProvider =
    Provider<ApiClient>((ref) => ApiClient(ref.watch(sessionStoreProvider)));
final authControllerProvider =
    AsyncNotifierProvider<AuthController, Account?>(AuthController.new);

class AuthController extends AsyncNotifier<Account?> {
  ApiClient get _api => ref.read(apiClientProvider);
  SessionStore get _store => ref.read(sessionStoreProvider);

  @override
  Future<Account?> build() async {
    if (await _store.read() == null) return null;
    try {
      final data = await _api.getJson('/api/v1/app/auth/me');
      return Account.fromJson(data);
    } catch (_) {
      await _store.clear();
      return null;
    }
  }

  Future<void> login(String username, String password) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final data = await _api.postJson('/api/v1/app/auth/login', {
        'username': username,
        'password': password,
        'deviceName': 'Flutter App'
      });
      await _saveTokens(data);
      return Account.fromJson(data['user'] as Map<String, dynamic>);
    });
  }

  Future<void> register(
      {required String username,
      required String password,
      required String role,
      String? companyName,
      String? displayName,
      String? contactName,
      String? phone}) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final data = await _api.postJson('/api/v1/app/auth/register', {
        'username': username,
        'password': password,
        'role': role,
        'companyName': companyName,
        'displayName': displayName,
        'contactName': contactName,
        'phone': phone,
        'deviceName': 'Flutter App'
      });
      await _saveTokens(data);
      return Account.fromJson(data['user'] as Map<String, dynamic>);
    });
  }

  Future<void> logout() async {
    try {
      await _api.postJson('/api/v1/app/auth/logout', const {});
    } finally {
      await _store.clear();
      state = const AsyncData(null);
    }
  }

  Future<void> _saveTokens(Map<String, dynamic> data) =>
      _store.write(SessionTokens(
          accessToken: data['accessToken'] as String,
          refreshToken: data['refreshToken'] as String));
}
