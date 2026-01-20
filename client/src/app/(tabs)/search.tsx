import React, { useState } from 'react';
import { 
  StyleSheet, 
  View, 
  TextInput, 
  TouchableOpacity, 
  ActivityIndicator, 
  Text, 
  SafeAreaView, 
  StatusBar,
  Keyboard,
  Platform
} from 'react-native';
import { COLORS, API_URL } from '@/src/constants/Config';
import { IconSymbol } from '@/src/components/ui/icon-symbol';
import { LeaderboardItem } from '@/src/components/LeaderboardItem';
import { ThemedText } from '@/src/components/themed-text';

interface SearchResult {
  id: number;
  username: string;
  rating: number;
  rank: number;
}

export default function SearchScreen() {
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<SearchResult | null>(null);
  const [error, setError] = useState('');

  const handleSearch = async () => {
    if (!query.trim()) return;
    
    setLoading(true);
    setError('');
    Keyboard.dismiss();

    try {
      const response = await fetch(`${API_URL}/users/rank?username=${query.trim()}`);
      if (!response.ok) {
        if (response.status === 404) throw new Error('User not found');
        throw new Error('Search failed');
      }
      const data = await response.json();
      setResult(data);
    } catch (err: any) {
      setError(err.message || 'Something went wrong');
      setResult(null);
    } finally {
      setLoading(false);
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <StatusBar barStyle="light-content" />
      <View style={styles.header}>
        <ThemedText type="title" style={styles.title}>Search User</ThemedText>
        <Text style={styles.subtitle}>Find global rank by username</Text>
      </View>

      <View style={styles.searchContainer}>
        <View style={styles.inputWrapper}>
          <IconSymbol name="magnifyingglass" size={20} color={COLORS.textMuted} style={styles.searchIcon} />
          <TextInput
            style={styles.input}
            placeholder="Enter username (e.g. user_00001)"
            placeholderTextColor={COLORS.textMuted}
            value={query}
            onChangeText={setQuery}
            onSubmitEditing={handleSearch}
            autoCapitalize="none"
            autoCorrect={false}
          />
        </View>
        <TouchableOpacity 
          style={styles.searchButton} 
          onPress={handleSearch}
          disabled={loading || !query.trim()}
        >
          {loading ? (
            <ActivityIndicator color={COLORS.white} />
          ) : (
            <Text style={styles.searchButtonText}>Search</Text>
          )}
        </TouchableOpacity>
      </View>

      <View style={styles.content}>
        {error ? (
          <View style={styles.errorContainer}>
            <IconSymbol name="exclamationmark.triangle.fill" size={40} color={COLORS.accent} />
            <Text style={styles.errorText}>{error}</Text>
          </View>
        ) : result ? (
          <View style={styles.resultWrapper}>
            <Text style={styles.sectionTitle}>Search Result</Text>
            <LeaderboardItem 
              user={{ 
                ID: result.id, 
                Username: result.username, 
                Rating: result.rating 
              }} 
              rank={result.rank} 
            />
            <View style={styles.statsCard}>
              <View style={styles.statItem}>
                <Text style={styles.statLabel}>Status</Text>
                <Text style={[styles.statValue, { color: '#10b981' }]}>Active Player</Text>
              </View>
              <View style={styles.divider} />
              <View style={styles.statItem}>
                <Text style={styles.statLabel}>Level</Text>
                <Text style={styles.statValue}>Pro Tier</Text>
              </View>
            </View>
          </View>
        ) : (
          <View style={styles.placeholderContainer}>
            <IconSymbol name="person.fill.viewfinder" size={60} color={COLORS.surfaceLight} />
            <Text style={styles.placeholderText}>Search for a player to see their rank</Text>
          </View>
        )}
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: COLORS.background,
    paddingTop: Platform.OS === 'android' ? StatusBar.currentHeight : 0,
  },
  header: {
    paddingHorizontal: 20,
    paddingVertical: 20,
  },
  title: {
    color: COLORS.text,
    fontSize: 28,
  },
  subtitle: {
    color: COLORS.textMuted,
    fontSize: 14,
    marginTop: -4,
  },
  searchContainer: {
    paddingHorizontal: 20,
    flexDirection: 'row',
    gap: 12,
    marginBottom: 24,
  },
  inputWrapper: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: COLORS.background,
    borderRadius: 0,
    paddingHorizontal: 12,
    borderWidth: 1,
    borderColor: COLORS.surfaceLight,
  },
  searchIcon: {
    marginRight: 8,
  },
  input: {
    flex: 1,
    height: 50,
    color: COLORS.text,
    fontSize: 16,
  },
  searchButton: {
    backgroundColor: COLORS.text,
    borderRadius: 0,
    paddingHorizontal: 20,
    justifyContent: 'center',
    alignItems: 'center',
  },
  searchButtonText: {
    color: COLORS.background,
    fontWeight: 'bold',
    fontSize: 14,
    textTransform: 'uppercase',
  },
  content: {
    flex: 1,
    paddingHorizontal: 0,
  },
  sectionTitle: {
    color: COLORS.text,
    fontSize: 14,
    fontWeight: 'bold',
    marginBottom: 16,
    textTransform: 'uppercase',
    letterSpacing: 1,
    paddingHorizontal: 20,
  },
  resultWrapper: {
    marginTop: 10,
  },
  statsCard: {
    backgroundColor: COLORS.surface,
    borderRadius: 0,
    padding: 20,
    marginTop: 0,
    flexDirection: 'row',
    borderWidth: 1,
    borderColor: '#222',
  },
  statItem: {
    flex: 1,
    alignItems: 'center',
  },
  statLabel: {
    color: COLORS.textMuted,
    fontSize: 10,
    marginBottom: 4,
    textTransform: 'uppercase',
  },
  statValue: {
    color: COLORS.text,
    fontSize: 14,
    fontWeight: 'bold',
  },
  divider: {
    width: 1,
    backgroundColor: '#333',
    marginHorizontal: 10,
  },
  placeholderContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    opacity: 0.5,
  },
  placeholderText: {
    color: COLORS.textMuted,
    fontSize: 16,
    textAlign: 'center',
    marginTop: 16,
    maxWidth: '80%',
  },
  errorContainer: {
    marginTop: 50,
    alignItems: 'center',
  },
  errorText: {
    color: COLORS.accent,
    fontSize: 16,
    marginTop: 12,
  },
});
