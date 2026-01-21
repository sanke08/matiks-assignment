import React, { useState, useCallback, useRef } from 'react';
import { 
  StyleSheet, 
  View, 
  FlatList, 
  ActivityIndicator, 
  Text, 
  StatusBar,
  RefreshControl,
  Platform
} from 'react-native';
import { useFocusEffect } from '@react-navigation/native';
import { COLORS, API_URL } from '@/src/constants/Config';
import { LeaderboardItem } from '@/src/components/LeaderboardItem';
import { ThemedText } from '@/src/components/themed-text';
import { SafeAreaView } from 'react-native-safe-area-context';

interface User {
  ID: number;
  Username: string;
  Rating: number;
  rank: number;
}

const PAGE_SIZE = 50;

export default function LeaderboardScreen() {
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [users, setUsers] = useState<User[]>([]);
  const [hasMore, setHasMore] = useState(true);
  
  const offsetRef = useRef(0);

  const fetchLeaderboard = async (isPoll = false, isRefresh = false) => {
    if (loadingMore && !isPoll && !isRefresh) return;
    
    const currentOffset = isRefresh ? 0 : offsetRef.current;
    
    try {
      const response = await fetch(`${API_URL}/leaderboard?limit=${PAGE_SIZE}&offset=${currentOffset}`);
      const data: User[] = await response.json();
      
      if (isRefresh || (isPoll && currentOffset === 0)) {
        // For poll/refresh of the first page, replace or merge carefully
        // Here we replace to ensure rank consistency
        setUsers(data || []);
        if (isRefresh) {
          offsetRef.current = 0;
          setHasMore(true);
        }
      } else {
        // For pagination, append
        setUsers(prev => [...prev, ...(data || [])]);
        if (!data || data.length < PAGE_SIZE) {
          setHasMore(false);
        }
      }
    } catch (error) {
      console.error('Error fetching leaderboard:', error);
    } finally {
      setLoading(false);
      setRefreshing(false);
      setLoadingMore(false);
    }
  };

  // Focus-aware polling for the FIRST PAGE ONLY to reduce costs
  useFocusEffect(
    useCallback(() => {
      // Initial load
      if (users.length === 0) fetchLeaderboard();

      const interval = setInterval(() => {
        // Only poll if we are at the top (Page 1) to save cost 
        // and prevent confusing the user while they are scrolling deep
        if (offsetRef.current === 0) {
          fetchLeaderboard(true);
        }
      }, 5000); // 5s interval is better for 10k users than 2s

      return () => clearInterval(interval);
    }, [users.length])
  );

  const onRefresh = useCallback(() => {
    setRefreshing(true);
    fetchLeaderboard(false, true);
  }, []);

  const loadMore = () => {
    if (!hasMore || loadingMore || loading) return;
    
    setLoadingMore(true);
    offsetRef.current += PAGE_SIZE;
    fetchLeaderboard();
  };

  if (loading && !refreshing) {
    return (
      <View style={styles.centered}>
        <ActivityIndicator size="large" color={COLORS.primary} />
        <Text style={styles.loadingText}>Loading ranks...</Text>
      </View>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <StatusBar barStyle="light-content" />
      <View style={styles.mainWrapper}>
        <View style={styles.header}>
          <View>
            <ThemedText type="title" style={styles.title}>Global Ranks</ThemedText>
            <Text style={styles.subtitle}>Top Performers</Text>
          </View>
        </View>

        <FlatList
          data={users}
          keyExtractor={(item, index) => `${item.ID}-${index}`}
          renderItem={({ item }) => (
            <LeaderboardItem user={item} rank={item.rank} />
          )}
          contentContainerStyle={styles.listContent}
          showsVerticalScrollIndicator={false}
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
          ListEmptyComponent={
            <View style={styles.emptyContainer}>
              <Text style={styles.emptyText}>No users found in leaderboard.</Text>
            </View>
          }
          ListFooterComponent={
            loadingMore ? (
              <View style={styles.footerLoader}>
                <ActivityIndicator color={COLORS.primary} />
              </View>
            ) : null
          }
        />
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
    backgroundColor: COLORS.surface, // Slightly different color for the content wrapper on wide screens
  },
  centered: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: COLORS.background,
  },
  loadingText: {
    color: COLORS.textMuted,
    marginTop: 12,
    fontSize: 16,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingVertical: 20,
    borderBottomWidth: 1,
    borderBottomColor: '#222',
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
  listContent: {
    paddingHorizontal: 0, 
    paddingBottom: 40,
  },
  footerLoader: {
    paddingVertical: 20,
    alignItems: 'center',
  },
  emptyContainer: {
    marginTop: 100,
    alignItems: 'center',
  },
  emptyText: {
    color: COLORS.textMuted,
    fontSize: 16,
  },
});
