import React from 'react'
import { View, ActivityIndicator, StyleSheet } from 'react-native'
import { useAuth } from '@/presentation/context/AuthContext'
import { AuthNavigator } from './AuthNavigator'
import { MainNavigator } from './MainNavigator'

export function RootNavigator() {
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return (
      <View style={styles.loading} testID="loading-indicator">
        <ActivityIndicator size="large" color="#2563eb" />
      </View>
    )
  }

  if (!isAuthenticated) {
    return <AuthNavigator />
  }

  return <MainNavigator />
}

const styles = StyleSheet.create({
  loading: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#fff',
  },
})
