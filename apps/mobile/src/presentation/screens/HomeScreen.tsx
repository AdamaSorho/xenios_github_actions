import { View, Text, StyleSheet } from 'react-native'

export function HomeScreen() {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Xenios</Text>
      <Text style={styles.subtitle}>Welcome to Xenios</Text>
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 20,
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    marginBottom: 10,
  },
  subtitle: {
    fontSize: 16,
    color: '#666',
  },
})
