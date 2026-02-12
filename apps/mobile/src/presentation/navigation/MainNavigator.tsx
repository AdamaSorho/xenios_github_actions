import React, { useState } from 'react'
import { View, TouchableOpacity, Text, StyleSheet } from 'react-native'
import { DashboardScreen } from '@/presentation/screens/DashboardScreen'
import { ClientsScreen } from '@/presentation/screens/ClientsScreen'
import { SessionsScreen } from '@/presentation/screens/SessionsScreen'
import { ProfileScreen } from '@/presentation/screens/ProfileScreen'

type MainTab = 'dashboard' | 'clients' | 'sessions' | 'profile'

const TABS: { key: MainTab; label: string }[] = [
  { key: 'dashboard', label: 'Home' },
  { key: 'clients', label: 'Clients' },
  { key: 'sessions', label: 'Sessions' },
  { key: 'profile', label: 'Profile' },
]

export function MainNavigator() {
  const [activeTab, setActiveTab] = useState<MainTab>('dashboard')

  const renderScreen = () => {
    switch (activeTab) {
      case 'dashboard':
        return <DashboardScreen />
      case 'clients':
        return <ClientsScreen />
      case 'sessions':
        return <SessionsScreen />
      case 'profile':
        return <ProfileScreen />
    }
  }

  return (
    <View style={styles.container}>
      <View style={styles.content}>{renderScreen()}</View>
      <View style={styles.tabBar}>
        {TABS.map((tab) => (
          <TouchableOpacity
            key={tab.key}
            style={[
              styles.tab,
              activeTab === tab.key && styles.tabActive,
            ]}
            onPress={() => setActiveTab(tab.key)}
            testID={`tab-${tab.key}`}
          >
            <Text
              style={[
                styles.tabText,
                activeTab === tab.key && styles.tabTextActive,
              ]}
            >
              {tab.label}
            </Text>
          </TouchableOpacity>
        ))}
      </View>
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  content: {
    flex: 1,
  },
  tabBar: {
    flexDirection: 'row',
    backgroundColor: '#fff',
    borderTopWidth: 1,
    borderTopColor: '#e5e5e5',
    paddingBottom: 20,
    paddingTop: 8,
  },
  tab: {
    flex: 1,
    alignItems: 'center',
    paddingVertical: 8,
  },
  tabActive: {
    borderTopWidth: 2,
    borderTopColor: '#2563eb',
  },
  tabText: {
    fontSize: 12,
    color: '#999',
    fontWeight: '500',
  },
  tabTextActive: {
    color: '#2563eb',
    fontWeight: '600',
  },
})
