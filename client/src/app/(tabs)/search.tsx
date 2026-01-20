import React, { useCallback, useRef, useState } from 'react';
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
  FlatList,
  RefreshControl
} from 'react-native';
import { COLORS, API_URL } from '@/src/constants/Config';
import { IconSymbol } from '@/src/components/ui/icon-symbol';
import { LeaderboardItem } from '@/src/components/LeaderboardItem';
import { ThemedText } from '@/src/components/themed-text';


const PAGE_SIZE=50
interface SearchResult {
  ID: number;
  Username: string;
  Rating: number;
  rank: number;
}

export default function SearchScreen() {
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [hasMore, setHasMore] = useState(true);
  const [error, setError] = useState('');
  
  const offsetRef = useRef(0);

  const handleSearch = async (isRefresh = false, isLoadMore = false) => {
    if (!query.trim()) return;
    
    if (!isLoadMore) {
      if (!isRefresh) setLoading(true);
      setError('');
      offsetRef.current = 0;
      setHasMore(true);
    } else {
      setLoadingMore(true);
    }

    Keyboard.dismiss();

    try {
      const currentOffset = offsetRef.current;
      const response = await fetch(`${API_URL}/leaderboard?username=${query.trim()}&limit=${PAGE_SIZE}&offset=${currentOffset}`);
      if (!response.ok) {
        throw new Error('Search failed');
      }
      const data: SearchResult[] = await response.json();
      
      if (isLoadMore) {
        setResults(prev => [...prev, ...(data || [])]);
      } else {
        setResults(data || []);
      }

      if (!data || data.length < PAGE_SIZE) {
        setHasMore(false);
      }

      if (!isLoadMore && data && data.length === 0) {
        setError('No users found');
      }
    } catch (err: any) {
      setError(err.message || 'Something went wrong');
      if (!isLoadMore) setResults([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
      setLoadingMore(false);
    }
  };

  const onRefresh = useCallback(() => {
    setRefreshing(true);
    handleSearch(true);
  }, [query]);

  const loadMore = () => {
    if (!hasMore || loadingMore || loading || !query.trim()) return;
    offsetRef.current += PAGE_SIZE;
    handleSearch(false, true);
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
            placeholder="Search username (e.g. rahul)"
            placeholderTextColor={COLORS.textMuted}
            value={query}
            onChangeText={(txt) => {
              setQuery(txt);
              if (!txt.trim()) {
                setResults([]);
                setError('');
              }
            }}
            onSubmitEditing={() => handleSearch()}
            autoCapitalize="none"
            autoCorrect={false}
          />
        </View>
        <TouchableOpacity 
          style={styles.searchButton} 
          onPress={() => handleSearch()}
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
        ) : results.length > 0 ? (
          <FlatList
            data={results}
            keyExtractor={(item, index) => `${item.ID}-${index}`}
            ListHeaderComponent={<Text style={styles.sectionTitle}>Search Results ({results.length})</Text>}
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
            contentContainerStyle={styles.listContent}
            onEndReached={loadMore}
            onEndReachedThreshold={0.5}
            refreshControl={
              <RefreshControl
                refreshing={refreshing}
                onRefresh={onRefresh}
                tintColor={COLORS.primary}
                colors={[COLORS.primary]}
              />
            }
            ListFooterComponent={
              loadingMore ? (
                <View style={styles.footerLoader}>
                  <ActivityIndicator color={COLORS.primary} />
                </View>
              ) : null
            }
          />
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
  footerLoader: {
    paddingVertical: 20,
    alignItems: 'center',
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
  listContent: {
    paddingBottom: 40,
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
