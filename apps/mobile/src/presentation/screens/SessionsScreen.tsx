import React from 'react'
import { View, Text, StyleSheet } from 'react-native'

export function SessionsScreen() {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Sessions</Text>
      <Text style={styles.subtitle}>Your training sessions</Text>
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
    backgroundColor: '#fff',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#1a1a1a',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: '#666',
  },
})
