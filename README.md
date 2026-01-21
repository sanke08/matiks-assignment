# High-Performance Leaderboard System

A full-stack real-time leaderboard application built for scale, capable of handling millions of users with millisecond-latency rank retrievals.

## 🚀 Project Overview

The project consists of two main parts:
1.  **Client (Mobile/Web)**: A high-performance React Native (Expo) application responsible for displaying the leaderboard with infinite scrolling and efficient updates.
2.  **Server (Backend)**: A robust Go backend leveraging Redis Sorted Sets for O(log N) ranking operations and PostgreSQL for persistent storage.

## 🛠️ Tech Stack

### Client
-   **Framework**: [React Native](https://reactnative.dev/) with [Expo](https://expo.dev/)
-   **Language**: TypeScript
-   **Routing**: Expo Router
-   **List Engine**: standard `FlatList` with optimization
-   **Styling**: React Native StyleSheet (Performance focused)

### Server
-   **Language**: Go (Golang)
-   **Database (Primary)**: PostgreSQL (via GORM)
-   **Cache/Leaderboard Engine**: Redis (Sorted Sets)
-   **Router**: Standard `net/http` ServeMux
-   **Driver**: `go-redis/v9`

## ✨ Key Features

-   **Real-time Ranking**: Instant rank calculation using Redis.
-   **Smart Polling**: The client intelligently polls for updates only when the user is at the top of the list to save bandwidth.
-   **Optimized List Rendering**: Infinite scrolling with pagination support.
-   **Dual-Store Architecture**: Data is persistent in Postgres but served hot from Redis for performance.

## 🏁 Getting Started

### Prerequisites
-   Go 1.22+
-   Node.js & pnpm
-   Redis (running on localhost:6379)
-   PostgreSQL (running on localhost:5432)

### 1. Backend Setup (Server)

1.  Navigate to the server directory:
    ```bash
    cd server
    ```
2.  Install dependencies:
    ```bash
    go mod download
    ```
3.  Set up your `.env` file (ensure DB credentials are correct).
4.  Run the server:
    ```bash
    go run internal/cmd/server/main.go
    ```
    *The server will start on port defined in config (usually 8080).*

### 2. Frontend Setup (Client)

1.  Navigate to the client directory:
    ```bash
    cd client
    ```
2.  Install dependencies:
    ```bash
    pnpm install
    ```
3.  Start the Expo app:
    ```bash
    npx expo start
    ```
    *Scan the QR code with your phone or press `w` for web / `a` for Android Emulator.*

## 📐 Architecture Highlights

### "Smart Pooling" Client Strategy
The client implements a resource-efficient polling mechanism that only refreshes data when the user is viewing the **top 50** users (`offset=0`). If the user scrolls down, polling stops to prevent content jumps and save server resources.

### Redis Sorted Sets
The backend uses Redis `ZSET` (Sorted Sets) to handle leaderboard logic.
-   **Score**: User Rating
-   **Member**: `Username:ID`
This allows fetching the "Top N Users" and "My Rank" in practically constant time, even with millions of records.







# Deep Technical Analysis

## 📱 Client-Side Analysis (React Native / Expo)

### List Rendering Strategy
The client uses `FlatList`, React Native's core component for rendering large lists efficiently. 

**Why this method?**
-   **Virtualization**: `FlatList` only renders items currently visible on the screen (plus a small buffer). This is critical for performance when the leaderboard could potentially have thousands of entries.
-   **Memory Management**: It recycles views that scroll off-screen, preventing memory bloat.

### Preventing Re-renders
1.  **Component Stability**: The `LeaderboardItem` is a separate functional component. While not explicitly wrapped in `React.memo`, strict separate props (string/number primitives) help reduce unnecessary checks relative to passing whole objects/functions.
2.  **State Isolation**: The list data resides in `users` state. Changes to other states (like `loadingMore` or `refreshing`) do not mutate the `users` array reference unnecessarily, helping `FlatList` avoid full re-renders of all rows.
3.  **Key Extraction**: Unique keys (`${item.ID}-${index}`) ensure React can diff the list efficiently. (Note: Using `index` users is a fallback; `item.ID` is preferred for stability).

### "Smart Pooling" (Polling) Implementation
You implemented a smart polling strategy in `LeaderboardScreen`:

```typescript
// Focus-aware polling for the FIRST PAGE ONLY
useFocusEffect(
  useCallback(() => {
    const interval = setInterval(() => {
      // Only poll if we are at the top (Page 1)
      if (offsetRef.current === 0) {
        fetchLeaderboard(true);
      }
    }, 5000); // 5s interval
    return () => clearInterval(interval);
  }, [users.length])
);
```

**How it works:**
1.  **Condition Check**: It checks `offsetRef.current === 0`. This means polling **only** happens when the user is viewing the very top of the list.
2.  **Benefit**:
    -   **Prevents Jumps**: If a user is scrolling down (reading rank #100), a background update won't suddenly shift the list or change the order of items under their finger.
    -   **Efficiency**: Top ranks change most frequently and are most viewed. Lower ranks change less often, so polling them is wasteful.

---

## 🖥️ Server-Side Analysis (Go + Redis)

### Execution Flow
1.  **Request**: Enters `cmd/server/main.go`, routed via `net/http` ServeMux.
2.  **Handler**: `handlers` package parses input.
3.  **Service**: `services` package contains business logic (e.g., validating input).
4.  **Repository**: `repository` package handles data access strategies.

### Redis Architecture
The project uses a **Write-Through** or **Side-Cache** pattern where Redis is the primary engine for leaderboard logic.

**Why Redis?**
-   **Sorted Sets (`ZSET`)**: Redis is chosen specifically for its `ZSET` data structure.
    -   `ZADD key score member`: Adds a user with a score (Rating) in O(log N).
    -   `ZREVRANGE key start stop`: Fetches top users in O(log N + M).
    -   `ZRANK key member`: Finds a user's exact rank in O(log N).
-   **Performance**: Postgres would require `ORDER BY rating DESC LIMIT X`, which becomes slow (O(N log N) or O(N)) as the table grows to millions of rows. Redis handles this in milliseconds.

### Optimization: "Unique Score Counting"
In `GetLeaderboard`, an advanced optimization is used:
1.  **Fetch Top Users**: Gets the top 50 users from Redis.
2.  **Calculate Ranks Efficiently**: Instead of asking Redis for the rank of *each* user individually (50 queries), it:
    -   Identifies unique scores among those 50 users.
    -   Pipelines `ZCOUNT` requests for just those unique scores.
    -   This significantly reduces the number of round-trips to Redis, essentially implementing "Dense Ranking" logic efficiently.

### Data Storage Strategy
-   **Member Format**: `username:ID`.
-   **Why?**: By storing the username and ID directly in the Redis member string, the server **skips a database lookup**.
    -   *Standard way*: Get ID from Redis -> Query Postgres for Name -> Return.
    -   *Your way*: Get `Name:ID` from Redis -> Split string -> Return.
    -   **Result**: Zero separate DB queries for reading the leaderboard. Extremely fast.
