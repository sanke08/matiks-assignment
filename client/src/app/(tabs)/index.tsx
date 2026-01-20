import React, { useEffect, useState, useCallback } from 'react';
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
import { COLORS, API_URL } from '@/src/constants/Config';
import { LeaderboardItem } from '@/src/components/LeaderboardItem';
import { ThemedText } from '@/src/components/themed-text';
import { SafeAreaView } from 'react-native-safe-area-context';

interface User {
  ID: number;
  Username: string;
  Rating: number;
}

export default function LeaderboardScreen() {
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [users, setUsers] = useState<User[]>([]);
  const [isSimulating, setIsSimulating] = useState(false);

  const fetchLeaderboard = async () => {
    try {
      const response = await fetch(`${API_URL}/leaderboard?limit=50`);
      const data = await response.json();
      setUsers(data || []);
    } catch (error) {
      console.error('Error fetching leaderboard:', error);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const checkSimulationStatus = async () => {
    try {
      const response = await fetch(`${API_URL}/simulation/status`);
      const data = await response.json();
      setIsSimulating(data.status === 'running');
    } catch (error) {
      console.log('Error checking simulation status');
    }
  };

  useEffect(() => {
    fetchLeaderboard();
    checkSimulationStatus();

    // Poll for updates every 2 seconds to show "Live" updates
    const interval = setInterval(() => {
      fetchLeaderboard();
      checkSimulationStatus();
    }, 2000);

    return () => clearInterval(interval);
  }, []);

  const onRefresh = useCallback(() => {
    setRefreshing(true);
    fetchLeaderboard();
    checkSimulationStatus();
  }, []);

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
      <View style={styles.header}>
        <View>
          <ThemedText type="title" style={styles.title}>Global Ranks</ThemedText>
          <Text style={styles.subtitle}>Top Performers</Text>
        </View>
        {isSimulating && (
          <View style={styles.liveIndicator}>
            <View style={styles.liveDot} />
            <Text style={styles.liveText}>LIVE</Text>
          </View>
        )}
      </View>

      <FlatList
        data={users}
        keyExtractor={(item) => item.ID.toString()}
        renderItem={({ item, index }) => (
          <LeaderboardItem user={item} rank={index + 1} />
        )}
        contentContainerStyle={styles.listContent}
        showsVerticalScrollIndicator={false}
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
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: COLORS.background,
    paddingTop: Platform.OS === 'android' ? StatusBar.currentHeight : 0,
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
  liveIndicator: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#333',
    paddingHorizontal: 12,
    paddingVertical: 4,
    borderRadius: 0,
    borderWidth: 1,
    borderColor: COLORS.white,
  },
  liveDot: {
    width: 6,
    height: 6,
    borderRadius: 0,
    backgroundColor: COLORS.white,
    marginRight: 8,
  },
  liveText: {
    color: COLORS.white,
    fontSize: 11,
    fontWeight: 'bold',
    letterSpacing: 1,
  },
  listContent: {
    paddingHorizontal: 0, // Full width items
    paddingBottom: 40,
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
