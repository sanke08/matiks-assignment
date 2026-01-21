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
  Platform,
  FlatList
} from 'react-native';
import { COLORS, API_URL } from '@/src/constants/Config';
import { IconSymbol } from '@/src/components/ui/icon-symbol';
import { LeaderboardItem } from '@/src/components/LeaderboardItem';
import { ThemedText } from '@/src/components/themed-text';

interface SearchResult {
  ID: number;
  Username: string;
  Rating: number;
  rank: number;
}

export default function SearchScreen() {
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [hasSearched, setHasSearched] = useState(false);
  const [error, setError] = useState('');

  const handleSearch = async () => {
    if (!query.trim()) return;
    
    setLoading(true);
    setError('');
    setHasSearched(true);
    Keyboard.dismiss();

    try {
      const response = await fetch(`${API_URL}/users/rank?username=${query.trim()}`);
      if (!response.ok) {
        throw new Error('Search failed');
      }
      const data = await response.json();
      setResults(data || []);
    } catch (err: any) {
      setError(err.message || 'Something went wrong');
      setResults([]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <StatusBar barStyle="light-content" />
      <View style={styles.mainWrapper}>
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
          ) : hasSearched ? (
            <FlatList
              data={results}
              keyExtractor={(item) => item.ID.toString()}
              ListHeaderComponent={() => (
                <Text style={styles.sectionTitle}>Search Results ({results.length})</Text>
              )}
              renderItem={({ item }) => (
                <LeaderboardItem 
                  user={{ 
                    ID: item.ID, 
                    Username: item.Username, 
                    Rating: item.Rating 
                  }} 
                  rank={item.rank} 
                />
              )}
              ListEmptyComponent={() => (
                <View style={styles.placeholderContainer}>
                  <IconSymbol name="person.fill.questionmark" size={60} color={COLORS.surfaceLight} />
                  <Text style={styles.placeholderText}>No players found matching "{query}"</Text>
                </View>
              )}
            />
          ) : (
            <View style={styles.placeholderContainer}>
              <IconSymbol name="person.fill.viewfinder" size={60} color={COLORS.surfaceLight} />
              <Text style={styles.placeholderText}>Search for a player to see their rank</Text>
            </View>
          )}
        </View>
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
  mainWrapper: {
    flex: 1,
    width: '100%',
    maxWidth: 800,
    alignSelf: 'center',
    backgroundColor: COLORS.surface,
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
