import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { COLORS } from '../constants/Config';
import { IconSymbol } from './ui/icon-symbol';

interface User {
  ID: number;
  Username: string;
  Rating: number;
  Rank?: number; // Optional if we calculate it locally or use it from API
}

interface LeaderboardItemProps {
  user: User;
  rank: number;
}

export const LeaderboardItem = ({ user, rank }: LeaderboardItemProps) => {
  const getRankStyle = () => {
    if (rank === 1) return styles.rank1;
    if (rank === 2) return styles.rank2;
    if (rank === 3) return styles.rank3;
    return styles.rankDefault;
  };

  const getRankIcon = () => {
    if (rank === 1) return <IconSymbol name="crown.fill" size={16} color={COLORS.gold} />;
    return null;
  };

  return (
    <View style={styles.container}>
      <View style={[styles.rankContainer, getRankStyle()]}>
        <Text style={styles.rankText}>{rank}</Text>
      </View>
      
      <View style={styles.userInfo}>
        <View style={styles.usernameRow}>
          <Text style={styles.username}>{user.Username}</Text>
          {getRankIcon()}
        </View>
        <Text style={styles.userId}>ID: {user.ID}</Text>
      </View>

      <View style={styles.ratingContainer}>
        <Text style={styles.ratingLabel}>Rating</Text>
        <Text style={styles.ratingValue}>{user.Rating}</Text>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: COLORS.surface,
    padding: 16,
    borderRadius: 0, // Squared for a more industrial mono look
    marginBottom: 1, // Thin line between items
    borderWidth: 1,
    borderColor: '#222',
  },
  rankContainer: {
    width: 32,
    height: 32,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 16,
  },
  rank1: { borderBottomWidth: 2, borderBottomColor: COLORS.white },
  rank2: { borderBottomWidth: 1, borderBottomColor: COLORS.white },
  rank3: { borderBottomWidth: 1, borderBottomColor: COLORS.white },
  rankDefault: {},
  rankText: {
    color: COLORS.text,
    fontWeight: 'bold',
    fontSize: 16,
  },
  userInfo: {
    flex: 1,
  },
  usernameRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  username: {
    color: COLORS.text,
    fontSize: 18,
    fontWeight: '600',
  },
  userId: {
    color: COLORS.textMuted,
    fontSize: 12,
    marginTop: 2,
  },
  ratingContainer: {
    alignItems: 'flex-end',
  },
  ratingLabel: {
    color: COLORS.textMuted,
    fontSize: 10,
    textTransform: 'uppercase',
    letterSpacing: 1,
  },
  ratingValue: {
    color: COLORS.primary,
    fontSize: 20,
    fontWeight: 'bold',
  },
});
