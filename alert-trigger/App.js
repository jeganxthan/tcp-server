import 'react-native-get-random-values';
import 'react-native-url-polyfill/auto';
import React, { useState } from 'react';
import {
  StyleSheet,
  Text,
  View,
  TouchableOpacity,
  ScrollView,
  SafeAreaView,
  StatusBar,
  Dimensions,
  Platform,
} from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import * as Haptics from 'expo-haptics';
import {
  AlertTriangle,
  Car,
  Eye,
  ShieldAlert,
  Smartphone,
  UserCheck,
  Zap,
  Disc,
  Navigation,
  Wind,
  Video,
  Cigarette,
  MonitorOff,
  UserPlus,
  LogIn,
  BellRing
} from 'lucide-react-native';
import { Client, Functions } from 'react-native-appwrite';

const { width } = Dimensions.get('window');

// Appwrite Configuration
const APPWRITE_ENDPOINT = 'https://app.ridemap365.in/v1';
const PROJECT_ID = '6a08450000344e9c5409';
const FUNCTION_ID = '6a084540002ebebb8559';

const client = new Client()
  .setEndpoint(APPWRITE_ENDPOINT)
  .setProject(PROJECT_ID);

const functions = new Functions(client);

const AlertButton = ({ title, icon: Icon, color, onPress }) => (
  <TouchableOpacity
    activeOpacity={0.7}
    onPress={() => {
      Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);
      onPress();
    }}
    style={[styles.buttonContainer, { borderColor: color + '40' }]}
  >
    <LinearGradient
      colors={[color + '20', color + '05']}
      start={{ x: 0, y: 0 }}
      end={{ x: 1, y: 1 }}
      style={styles.buttonGradient}
    >
      <View style={[styles.iconWrapper, { backgroundColor: color + '30' }]}>
        <Icon size={24} color={color} />
      </View>
      <Text style={styles.buttonText}>{title}</Text>
    </LinearGradient>
  </TouchableOpacity>
);

const Section = ({ title, children }) => (
  <View style={styles.section}>
    <Text style={styles.sectionTitle}>{title}</Text>
    <View style={styles.grid}>{children}</View>
  </View>
);

export default function App() {
  const [lastAction, setLastAction] = useState(null);

  const triggerAlert = async (path, label) => {
    try {
      setLastAction(`Triggering ${label}...`);

      const execution = await functions.createExecution(
        FUNCTION_ID,
        '', // body not needed for this routing style
        false, // async
        path, // path
        'GET' // method
      );

      const response = JSON.parse(execution.responseBody);
      if (response.status === 'error') {
        setLastAction(`${label} Fail: ${response.message}`);
      } else {
        setLastAction(`${label}: ${response.status || 'Success'}`);
      }
      setTimeout(() => setLastAction(null), 5000);
    } catch (error) {
      console.error(error);
      setLastAction(`Error: ${error.message}`);
      setTimeout(() => setLastAction(null), 6000);
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <StatusBar barStyle="light-content" />
      <LinearGradient
        colors={['#0f172a', '#1e293b', '#0f172a']}
        style={StyleSheet.absoluteFill}
      />

      <ScrollView contentContainerStyle={styles.scrollContent}>
        <View style={styles.header}>
          <Text style={styles.title}>ThirdEye AI</Text>
          <Text style={styles.subtitle}>Alert Control Center (Appwrite)</Text>
        </View>

        {lastAction && (
          <View style={styles.statusToast}>
            <BellRing size={16} color="#fbbf24" style={{ marginRight: 8 }} />
            <Text style={styles.statusText}>{lastAction}</Text>
          </View>
        )}

        <Section title="ADAS ALERTS">
          <AlertButton
            title="Pedestrian"
            icon={UserCheck}
            color="#ef4444"
            onPress={() => triggerAlert('/alert/pedestrian', 'Pedestrian Alert')}
          />
          <AlertButton
            title="Theft Alert"
            icon={ShieldAlert}
            color="#f97316"
            onPress={() => triggerAlert('/alert/theft', 'Theft Alert')}
          />
          <AlertButton
            title="Collision"
            icon={Car}
            color="#dc2626"
            onPress={() => triggerAlert('/alert/forward-collision', 'Forward Collision')}
          />
        </Section>

        <Section title="DMS ALERTS">
          <AlertButton
            title="Drowsy"
            icon={Eye}
            color="#ef4444"
            onPress={() => triggerAlert('/alert/drowsy', 'Drowsiness')}
          />
          <AlertButton
            title="Distraction"
            icon={AlertTriangle}
            color="#f59e0b"
            onPress={() => triggerAlert('/alert/distraction', 'Distraction')}
          />
          <AlertButton
            title="Mobile Use"
            icon={Smartphone}
            color="#f59e0b"
            onPress={() => triggerAlert('/alert/mobile', 'Mobile Usage')}
          />
          <AlertButton
            title="No Helmet"
            icon={ShieldAlert}
            color="#f97316"
            onPress={() => triggerAlert('/alert/no-helmet', 'No Helmet')}
          />
          <AlertButton
            title="Cam Hidden"
            icon={MonitorOff}
            color="#ef4444"
            onPress={() => triggerAlert('/alert/camera-hidden', 'Camera Hidden')}
          />
          <AlertButton
            title="Smoking"
            icon={Cigarette}
            color="#8b5cf6"
            onPress={() => triggerAlert('/alert/smoking', 'Smoking')}
          />
          <AlertButton
            title="Driver OK"
            icon={UserCheck}
            color="#10b981"
            onPress={() => triggerAlert('/alert/driver-detected', 'Driver Detected')}
          />
        </Section>

        <Section title="VEHICLE DYNAMICS">
          <AlertButton
            title="Speeding"
            icon={Zap}
            color="#ef4444"
            onPress={() => triggerAlert('/alert/speed', 'Speed Alert')}
          />
          <AlertButton
            title="Harsh Brake"
            icon={Disc}
            color="#f59e0b"
            onPress={() => triggerAlert('/alert/harsh-brake', 'Harsh Brake')}
          />
          <AlertButton
            title="Harsh Accel"
            icon={Wind}
            color="#3b82f6"
            onPress={() => triggerAlert('/alert/harsh-acceleration', 'Harsh Accel')}
          />
          <AlertButton
            title="Sharp Turn"
            icon={Navigation}
            color="#8b5cf6"
            onPress={() => triggerAlert('/alert/sharp-turn', 'Sharp Turn')}
          />
        </Section>

        <Section title="DRIVER MANAGEMENT">
          <AlertButton
            title="Login"
            icon={LogIn}
            color="#10b981"
            onPress={() => triggerAlert('/login', 'Login')}
          />
          <AlertButton
            title="Register"
            icon={UserPlus}
            color="#6366f1"
            onPress={() => triggerAlert('/register', 'Register')}
          />
        </Section>

        <View style={styles.footer}>
          <Text style={styles.footerText}>Appwrite: {APPWRITE_ENDPOINT}</Text>
          <Text style={styles.footerText}>Project: {PROJECT_ID}</Text>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0f172a',
  },
  scrollContent: {
    padding: 20,
    paddingTop: Platform.OS === 'android' ? 40 : 20,
  },
  header: {
    marginBottom: 30,
  },
  title: {
    fontSize: 32,
    fontWeight: '800',
    color: '#f8fafc',
    letterSpacing: -1,
  },
  subtitle: {
    fontSize: 16,
    color: '#94a3b8',
    marginTop: 4,
  },
  statusToast: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#1e293b',
    padding: 12,
    borderRadius: 12,
    marginBottom: 20,
    borderWidth: 1,
    borderColor: '#334155',
  },
  statusText: {
    color: '#fbbf24',
    fontWeight: '600',
    fontSize: 14,
  },
  section: {
    marginBottom: 32,
  },
  sectionTitle: {
    fontSize: 14,
    fontWeight: '700',
    color: '#64748b',
    textTransform: 'uppercase',
    letterSpacing: 1.5,
    marginBottom: 16,
    marginLeft: 4,
  },
  grid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
  },
  buttonContainer: {
    width: (width - 56) / 2,
    height: 100,
    borderRadius: 20,
    marginBottom: 16,
    borderWidth: 1.5,
    overflow: 'hidden',
  },
  buttonGradient: {
    flex: 1,
    padding: 16,
    justifyContent: 'center',
    alignItems: 'center',
  },
  iconWrapper: {
    width: 44,
    height: 44,
    borderRadius: 14,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 8,
  },
  buttonText: {
    color: '#f1f5f9',
    fontSize: 14,
    fontWeight: '600',
    textAlign: 'center',
  },
  footer: {
    marginTop: 20,
    paddingBottom: 40,
    alignItems: 'center',
  },
  footerText: {
    color: '#475569',
    fontSize: 10,
    fontFamily: Platform.OS === 'ios' ? 'Menlo' : 'monospace',
  },
});
