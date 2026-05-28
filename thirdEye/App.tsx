import React, { useState, useEffect, useRef } from 'react';
import {
  View,
  Text,
  StyleSheet,
  Animated,
  Easing,
  TouchableOpacity,
  ScrollView,
  Platform,
  StatusBar,
  Dimensions,
  ActivityIndicator
} from 'react-native';
import { AlertTriangle, Info, ShieldAlert, CheckCircle, Car, Activity, Coffee, UserCheck, ShieldCheck, Camera, Eye } from 'lucide-react-native';
import { SafeAreaView, SafeAreaProvider } from 'react-native-safe-area-context';
import { LinearGradient } from 'expo-linear-gradient';
import { CameraView, useCameraPermissions } from 'expo-camera';
import { Client, Databases } from 'react-native-appwrite';
import 'react-native-get-random-values';
import 'react-native-url-polyfill/auto';

const { width, height } = Dimensions.get('window');

interface AlertData {
  type: string;
  message: string;
  severity: string;
  timestamp: string;
}

const severityConfig: Record<string, { color: string, bg: string, border: string, icon: any }> = {
  CRITICAL: { color: '#ef4444', bg: 'rgba(239, 68, 68, 0.15)', border: 'rgba(239, 68, 68, 0.3)', icon: ShieldAlert },
  HIGH: { color: '#f97316', bg: 'rgba(249, 115, 22, 0.15)', border: 'rgba(249, 115, 22, 0.3)', icon: AlertTriangle },
  MEDIUM: { color: '#eab308', bg: 'rgba(234, 179, 8, 0.15)', border: 'rgba(234, 179, 8, 0.3)', icon: Info },
  LOW: { color: '#3b82f6', bg: 'rgba(59, 130, 246, 0.15)', border: 'rgba(59, 130, 246, 0.3)', icon: CheckCircle },
};

// Appwrite Configuration
const APPWRITE_ENDPOINT = 'https://app.ridemap365.in/v1';
const PROJECT_ID = '6a08450000344e9c5409';
const DATABASE_ID = 'thirdeye_db';
const COLLECTION_ID = 'alerts';

const client = new Client()
  .setEndpoint(APPWRITE_ENDPOINT)
  .setProject(PROJECT_ID);

const databases = new Databases(client);

export default function App() {
  const [isRegistered, setIsRegistered] = useState(false);

  return (
    <SafeAreaProvider>
      <StatusBar barStyle="light-content" />
      {!isRegistered ? (
        <RegistrationScreen onComplete={() => setIsRegistered(true)} />
      ) : (
        <ThirdEyeApp />
      )}
    </SafeAreaProvider>
  );
}

function RegistrationScreen({ onComplete }: { onComplete: () => void }) {
  const [status, setStatus] = useState<'idle' | 'scanning' | 'analyzing' | 'success'>('idle');
  const [permission, requestPermission] = useCameraPermissions();
  const scanLinePos = useRef(new Animated.Value(0)).current;
  const analysisScale = useRef(new Animated.Value(0)).current;
  const opacity = useRef(new Animated.Value(1)).current;

  const [analysisSteps, setAnalysisSteps] = useState([
    { label: 'Facial Symmetry', value: '98%', done: false },
    { label: 'Iris Pattern', value: '100%', done: false },
    { label: 'Driver ID Match', value: 'Match', done: false },
  ]);

  const handleStart = async () => {
    if (!permission?.granted) {
      const res = await requestPermission();
      if (!res.granted) return;
    }
    setStatus('scanning');

    Animated.loop(
      Animated.sequence([
        Animated.timing(scanLinePos, { toValue: 240, duration: 2000, useNativeDriver: true }),
        Animated.timing(scanLinePos, { toValue: 0, duration: 2000, useNativeDriver: true }),
      ])
    ).start();

    setTimeout(() => {
      setStatus('analyzing');
      Animated.spring(analysisScale, { toValue: 1, friction: 8, useNativeDriver: true }).start();

      let step = 0;
      const interval = setInterval(() => {
        setAnalysisSteps(prev => {
          const next = [...prev];
          if (next[step]) next[step].done = true;
          return next;
        });
        step++;
        if (step > 2) {
          clearInterval(interval);
          setTimeout(() => {
            setStatus('success');
            setTimeout(() => {
              Animated.timing(opacity, { toValue: 0, duration: 800, useNativeDriver: true }).start(onComplete);
            }, 1000);
          }, 800);
        }
      }, 700);
    }, 3000);
  };

  return (
    <Animated.View style={[styles.regContainer, { opacity }]}>
      <LinearGradient colors={['#0f172a', '#1e1b4b']} style={StyleSheet.absoluteFillObject} />

      <View style={styles.regContent}>
        <Text style={styles.regTitle}>Saarathi AI</Text>
        <Text style={styles.regSub}>Biometric Driver Verification</Text>

        <View style={styles.scanWrapper}>
          <View style={styles.cameraBox}>
            {status !== 'idle' ? (
              <CameraView style={StyleSheet.absoluteFill} facing="front" />
            ) : (
              <View style={styles.facePlaceholder}><UserCheck color="#475569" size={80} /></View>
            )}

            {status === 'scanning' && <Animated.View style={[styles.scanLine, { transform: [{ translateY: scanLinePos }] }]} />}
            {status === 'analyzing' && <View style={styles.analysisOverlay}><ActivityIndicator color="#3b82f6" size="large" /><Text style={styles.analysisText}>ANALYZING...</Text></View>}
            {status === 'success' && <View style={[styles.analysisOverlay, { backgroundColor: 'rgba(16, 185, 129, 0.4)' }]}><ShieldCheck color="#fff" size={80} /><Text style={[styles.analysisText, { color: '#fff' }]}>VERIFIED</Text></View>}
          </View>
          <View style={[styles.corner, styles.topLeft]} /><View style={[styles.corner, styles.topRight]} /><View style={[styles.corner, styles.bottomLeft]} /><View style={[styles.corner, styles.bottomRight]} />
        </View>

        {status === 'analyzing' && (
          <Animated.View style={[styles.analysisList, { transform: [{ scale: analysisScale }] }]}>
            {analysisSteps.map((step, i) => (
              <View key={i} style={styles.analysisRow}>
                <Text style={styles.analysisLabel}>{step.label}</Text>
                <Text style={[styles.analysisValue, { color: step.done ? '#10b981' : '#64748b' }]}>{step.value}</Text>
              </View>
            ))}
          </Animated.View>
        )}

        {status === 'idle' && (
          <TouchableOpacity onPress={handleStart} style={styles.regBtn}>
            <Text style={styles.regBtnText}>START VERIFICATION</Text>
          </TouchableOpacity>
        )}
      </View>
    </Animated.View>
  );
}

function ThirdEyeApp() {
  const [isConnected, setIsConnected] = useState(false);
  const [alerts, setAlerts] = useState<AlertData[]>([]);
  const [currentAlert, setCurrentAlert] = useState<AlertData | null>(null);
  const [drowsinessCount, setDrowsinessCount] = useState(0);
  const [showTeaBreak, setShowTeaBreak] = useState(false);

  const alertOpacity = useRef(new Animated.Value(0)).current;
  const alertScale = useRef(new Animated.Value(0.8)).current;
  const bgOpacity = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    // Initial connection status
    setIsConnected(true);

    // Subscribe to real-time alerts
    const unsubscribe = client.subscribe(
      `databases.${DATABASE_ID}.collections.${COLLECTION_ID}.documents`,
      (response) => {
        if (response.events.includes('databases.*.collections.*.documents.*.create')) {
          const data = response.payload as AlertData;
          console.log('Appwrite Alert Received:', data);

          if (data.type === 'DROWSINESS') {
            setDrowsinessCount(prev => {
              const next = prev + 1;
              if (next >= 2) {
                setShowTeaBreak(true);
                return 0; // Reset
              }
              return next;
            });
          }

          setAlerts(prev => [data, ...prev].slice(0, 50));
          showBigAlert(data);
        }
      }
    );

    return () => unsubscribe();
  }, []);

  const showBigAlert = (alertData: AlertData) => {
    setCurrentAlert(alertData);
    Animated.parallel([
      Animated.timing(bgOpacity, { toValue: 1, duration: 200, useNativeDriver: true }),
      Animated.timing(alertOpacity, { toValue: 1, duration: 300, useNativeDriver: true }),
      Animated.spring(alertScale, { toValue: 1, friction: 6, useNativeDriver: true })
    ]).start();

    setTimeout(() => {
      Animated.timing(bgOpacity, { toValue: 0, duration: 300, useNativeDriver: true }).start(() => setCurrentAlert(null));
    }, 5000);
  };

  return (
    <SafeAreaView style={styles.container}>
      <LinearGradient colors={['#0f172a', '#1e293b']} style={StyleSheet.absoluteFillObject} />
      <View style={styles.header}>
        <View style={{ flexDirection: 'row', alignItems: 'center' }}>
          <View style={styles.logoContainer}><Car color="#3b82f6" size={24} /></View>
          <Text style={styles.headerTitle}>Saarathi Live</Text>
        </View>
        <View style={[styles.statusBadge, { borderColor: isConnected ? '#10b981' : '#ef4444' }]}>
          <View style={[styles.statusDot, { backgroundColor: isConnected ? '#10b981' : '#ef4444' }]} />
          <Text style={{ color: isConnected ? '#10b981' : '#ef4444', fontWeight: 'bold', fontSize: 10 }}>{isConnected ? 'SECURE' : 'OFFLINE'}</Text>
        </View>
      </View>

      <View style={{ flex: 1, padding: 20 }}>
        <Text style={styles.sectionTitle}>REAL-TIME LOG</Text>
        <ScrollView showsVerticalScrollIndicator={false}>
          {alerts.length === 0 ? (
            <Text style={{ color: '#475569', textAlign: 'center', marginTop: 40 }}>Waiting for telemetry...</Text>
          ) : (
            alerts.map((alert, i) => {
              const cfg = severityConfig[alert.severity] || severityConfig.LOW;
              const Icon = cfg.icon;
              return (
                <View key={i} style={[styles.historyCard, { borderColor: cfg.border }]}>
                  <View style={[styles.iconWrapper, { backgroundColor: cfg.bg }]}><Icon color={cfg.color} size={20} /></View>
                  <View style={{ flex: 1 }}>
                    <Text style={styles.historyMessage}>{alert.message}</Text>
                    <Text style={styles.historyTime}>{new Date(alert.timestamp).toLocaleTimeString()}</Text>
                  </View>
                  <Text style={{ color: cfg.color, fontWeight: 'bold', fontSize: 10 }}>{alert.severity}</Text>
                </View>
              );
            })
          )}
        </ScrollView>
      </View>

      {showTeaBreak && (
        <View style={styles.teaBreakOverlay}>
          <LinearGradient colors={['#7c3aed', '#4f46e5']} style={styles.teaBreakCard}>
            <Coffee color="#fff" size={60} />
            <Text style={styles.teaBreakTitle}>Take a break and drink some tea!</Text>
            <Text style={styles.teaBreakSub}>Drowsiness detected multiple times. Safety first.</Text>
            <TouchableOpacity onPress={() => setShowTeaBreak(false)} style={styles.teaBreakBtn}>
              <Text style={{ fontWeight: 'bold' }}>I'M REFRESHED</Text>
            </TouchableOpacity>
          </LinearGradient>
        </View>
      )}

      {currentAlert && !showTeaBreak && (
        <Animated.View style={[StyleSheet.absoluteFill, styles.overlay, { opacity: bgOpacity, backgroundColor: currentAlert.severity === 'CRITICAL' ? 'rgba(153, 27, 27, 0.95)' : 'rgba(15, 23, 42, 0.95)' }]}>
          <Animated.View style={[styles.popup, { opacity: alertOpacity, transform: [{ scale: alertScale }] }]}>
            {(() => { const Icon = severityConfig[currentAlert.severity]?.icon || AlertTriangle; return <Icon color="#fff" size={100} />; })()}
            <Text style={styles.popupMsg}>{currentAlert.message}</Text>
            <Text style={{ color: 'rgba(255,255,255,0.6)', fontWeight: 'bold', marginTop: 10, letterSpacing: 2 }}>{currentAlert.severity}</Text>

            <TouchableOpacity
              activeOpacity={0.8}
              onPress={() => dismissAlert()}
              style={styles.ackBtn}
            >
              <Text style={styles.ackBtnText}>ACKNOWLEDGE (5s)</Text>
            </TouchableOpacity>
          </Animated.View>
        </Animated.View>
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
  header: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', padding: 20, borderBottomWidth: 1, borderBottomColor: 'rgba(255,255,255,0.05)' },
  logoContainer: { backgroundColor: 'rgba(59, 130, 246, 0.1)', padding: 8, borderRadius: 12, marginRight: 10 },
  headerTitle: { color: '#fff', fontSize: 20, fontWeight: '900' },
  statusBadge: { flexDirection: 'row', alignItems: 'center', paddingHorizontal: 10, paddingVertical: 4, borderRadius: 12, borderWidth: 1 },
  statusDot: { width: 6, height: 6, borderRadius: 3, marginRight: 6 },
  sectionTitle: { color: '#64748b', fontWeight: 'bold', fontSize: 12, letterSpacing: 1.5, marginBottom: 15 },
  historyCard: { flexDirection: 'row', alignItems: 'center', padding: 14, marginBottom: 10, backgroundColor: 'rgba(30, 41, 59, 0.5)', borderRadius: 14, borderWidth: 1 },
  iconWrapper: { padding: 8, borderRadius: 10, marginRight: 12 },
  historyMessage: { color: '#fff', fontWeight: 'bold', fontSize: 14 },
  historyTime: { color: '#64748b', fontSize: 11, marginTop: 2 },
  overlay: { justifyContent: 'center', alignItems: 'center', zIndex: 100 },
  popup: { alignItems: 'center', width: '90%' },
  popupMsg: { color: '#fff', fontSize: 38, fontWeight: '900', textAlign: 'center', marginTop: 20 },
  regContainer: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  regContent: { width: '100%', alignItems: 'center', padding: 30 },
  regTitle: { color: '#fff', fontSize: 28, fontWeight: '900' },
  regSub: { color: '#64748b', fontSize: 14, marginBottom: 40 },
  scanWrapper: { width: 260, height: 260, justifyContent: 'center', alignItems: 'center' },
  cameraBox: { width: 240, height: 240, borderRadius: 120, overflow: 'hidden', backgroundColor: '#000', borderWidth: 2, borderColor: 'rgba(255,255,255,0.1)' },
  facePlaceholder: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  scanLine: { position: 'absolute', left: 0, right: 0, height: 3, backgroundColor: '#3b82f6', shadowColor: '#3b82f6', shadowRadius: 10, shadowOpacity: 1, elevation: 10 },
  analysisOverlay: { ...StyleSheet.absoluteFillObject, backgroundColor: 'rgba(15, 23, 42, 0.7)', justifyContent: 'center', alignItems: 'center' },
  analysisText: { color: '#3b82f6', fontWeight: 'bold', fontSize: 12, marginTop: 15 },
  analysisList: { width: '100%', marginTop: 30, backgroundColor: 'rgba(30, 41, 59, 0.5)', borderRadius: 20, padding: 20 },
  analysisRow: { flexDirection: 'row', justifyContent: 'space-between', paddingVertical: 10, borderBottomWidth: 1, borderBottomColor: 'rgba(255,255,255,0.05)' },
  analysisLabel: { color: '#94a3b8', fontSize: 13 },
  analysisValue: { fontWeight: 'bold', fontSize: 13 },
  corner: { position: 'absolute', width: 40, height: 40, borderColor: '#3b82f6' },
  topLeft: { top: 0, left: 0, borderTopWidth: 4, borderLeftWidth: 4, borderTopLeftRadius: 30 },
  topRight: { top: 0, right: 0, borderTopWidth: 4, borderRightWidth: 4, borderTopRightRadius: 30 },
  bottomLeft: { bottom: 0, left: 0, borderBottomWidth: 4, borderLeftWidth: 4, borderBottomLeftRadius: 30 },
  bottomRight: { bottom: 0, right: 0, borderBottomWidth: 4, borderRightWidth: 4, borderBottomRightRadius: 30 },
  regBtn: { backgroundColor: '#3b82f6', paddingHorizontal: 50, paddingVertical: 18, borderRadius: 15, marginTop: 40 },
  regBtnText: { color: '#fff', fontWeight: 'bold', letterSpacing: 1 },
  teaBreakOverlay: { ...StyleSheet.absoluteFillObject, backgroundColor: 'rgba(0,0,0,0.9)', justifyContent: 'center', alignItems: 'center', zIndex: 1000, padding: 30 },
  teaBreakCard: { width: '100%', padding: 40, borderRadius: 30, alignItems: 'center' },
  teaBreakTitle: { color: '#fff', fontSize: 24, fontWeight: '900', marginTop: 20, textAlign: 'center' },
  teaBreakSub: { color: 'rgba(255,255,255,0.8)', textAlign: 'center', marginTop: 10, fontSize: 14 },
  teaBreakBtn: { backgroundColor: '#fff', paddingHorizontal: 30, paddingVertical: 15, borderRadius: 15, marginTop: 30 },
  ackBtn: { backgroundColor: 'rgba(255,255,255,0.2)', paddingHorizontal: 40, paddingVertical: 15, borderRadius: 12, marginTop: 30, borderWidth: 1, borderColor: 'rgba(255,255,255,0.3)' },
  ackBtnText: { color: '#fff', fontWeight: 'bold', letterSpacing: 1 }
});
