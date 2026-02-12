import React from 'react'
import { View, Text, StyleSheet, TouchableOpacity } from 'react-native'
import { useAuth } from '@/presentation/context/AuthContext'

export function ProfileScreen() {
  const { user, logout } = useAuth()

  return (
    <View style={styles.container}>
      <View style={styles.avatar}>
        <Text style={styles.avatarText}>
          {user?.name?.charAt(0).toUpperCase() || '?'}
        </Text>
      </View>
      <Text style={styles.name} testID="profile-name">
        {user?.name || 'Unknown'}
      </Text>
      <Text style={styles.email} testID="profile-email">
        {user?.email || 'No email'}
      </Text>
      <Text style={styles.role} testID="profile-role">
        {user?.role || 'Unknown'}
      </Text>
      <TouchableOpacity
        style={styles.logoutButton}
        onPress={logout}
        testID="logout-button"
      >
        <Text style={styles.logoutText}>Sign Out</Text>
      </TouchableOpacity>
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
  avatar: {
    width: 80,
    height: 80,
    borderRadius: 40,
    backgroundColor: '#2563eb',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 16,
  },
  avatarText: {
    color: '#fff',
    fontSize: 32,
    fontWeight: 'bold',
  },
  name: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#1a1a1a',
    marginBottom: 4,
  },
  email: {
    fontSize: 16,
    color: '#666',
    marginBottom: 4,
  },
  role: {
    fontSize: 14,
    color: '#999',
    textTransform: 'capitalize',
    marginBottom: 32,
  },
  logoutButton: {
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#dc2626',
  },
  logoutText: {
    color: '#dc2626',
    fontSize: 16,
    fontWeight: '500',
  },
})
