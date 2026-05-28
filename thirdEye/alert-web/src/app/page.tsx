"use client";

import React, { useState, useEffect } from 'react';
import { 
  Bell, 
  Shield, 
  Users, 
  Activity, 
  AlertTriangle, 
  Navigation, 
  BarChart3, 
  Settings, 
  Search,
  ChevronRight,
  Wifi,
  MapPin,
  TrendingUp,
  Circle,
  Eye
} from 'lucide-react';
import { 
  LineChart, 
  Line, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  AreaChart,
  Area
} from 'recharts';
import { motion, AnimatePresence } from 'framer-motion';

// Mock Analytics Data
const analyticsData = [
  { name: '00:00', alerts: 4, score: 98 },
  { name: '04:00', alerts: 2, score: 99 },
  { name: '08:00', alerts: 12, score: 85 },
  { name: '12:00', alerts: 25, score: 72 },
  { name: '16:00', alerts: 18, score: 78 },
  { name: '20:00', alerts: 8, score: 92 },
];

const mockDrivers = [
  { id: 1, name: 'Alex Johnson', status: 'Active', location: 'Downtown', risk: 'Low' },
  { id: 2, name: 'Sarah Miller', status: 'Warning', location: 'Highway 101', risk: 'High' },
  { id: 3, name: 'David Chen', status: 'Active', location: 'Park Ave', risk: 'Low' },
];

interface Alert {
  type: string;
  message: string;
  severity: string;
  timestamp: string;
}

export default function Dashboard() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [activeTab, setActiveTab] = useState('overview');
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    // Connect to the same Go backend as the mobile app
    const ws = new WebSocket('ws://139.59.73.32:5023/ws');
    
    ws.onopen = () => setIsConnected(true);
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setAlerts(prev => [data, ...prev].slice(0, 20));
    };
    ws.onclose = () => setIsConnected(false);
    
    return () => ws.close();
  }, []);

  return (
    <div className="flex h-screen bg-[#0f172a] text-slate-100 overflow-hidden">
      {/* Sidebar */}
      <aside className="w-64 border-r border-slate-800 flex flex-col p-6 bg-[#0f172a]">
        <div className="flex items-center gap-3 mb-10 px-2">
          <div className="bg-blue-600 p-2 rounded-xl">
            <Shield className="w-6 h-6 text-white" />
          </div>
          <h1 className="text-xl font-bold tracking-tight">Saarathi Web</h1>
        </div>

        <nav className="flex-1 space-y-2">
          {[
            { id: 'overview', icon: BarChart3, label: 'Overview' },
            { id: 'tracking', icon: Navigation, label: 'Live Tracking' },
            { id: 'drivers', icon: Users, label: 'Drivers' },
            { id: 'alerts', icon: AlertTriangle, label: 'Alert History' },
          ].map((item) => (
            <button
              key={item.id}
              onClick={() => setActiveTab(item.id)}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 ${
                activeTab === item.id 
                ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20' 
                : 'text-slate-400 hover:bg-slate-800 hover:text-slate-100'
              }`}
            >
              <item.icon className="w-5 h-5" />
              <span className="font-medium">{item.label}</span>
            </button>
          ))}
        </nav>

        <div className="mt-auto pt-6 border-t border-slate-800 space-y-2">
          <button className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-slate-400 hover:bg-slate-800 hover:text-slate-100 transition-all">
            <Settings className="w-5 h-5" />
            <span className="font-medium">Settings</span>
          </button>
          <div className="px-4 py-3 flex items-center gap-3">
            <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-emerald-500' : 'bg-rose-500'} animate-pulse`} />
            <span className="text-xs font-semibold text-slate-500 uppercase tracking-widest">
              {isConnected ? 'Backend Live' : 'Backend Offline'}
            </span>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="h-20 border-b border-slate-800 flex items-center justify-between px-8 bg-[#0f172a]/50 backdrop-blur-xl sticky top-0 z-10">
          <div className="relative w-96">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
            <input 
              type="text" 
              placeholder="Search driver or location..." 
              className="w-full bg-slate-900 border border-slate-800 rounded-xl py-2 pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50 transition-all"
            />
          </div>
          
          <div className="flex items-center gap-6">
            <div className="relative">
              <Bell className="w-5 h-5 text-slate-400 cursor-pointer hover:text-white transition-colors" />
              {alerts.length > 0 && (
                <span className="absolute -top-1 -right-1 w-2 h-2 bg-rose-500 rounded-full" />
              )}
            </div>
            <div className="h-8 w-px bg-slate-800" />
            <div className="flex items-center gap-3">
              <div className="text-right">
                <p className="text-sm font-semibold">Fleet Admin</p>
                <p className="text-xs text-slate-500">Administrator</p>
              </div>
              <div className="w-10 h-10 rounded-xl bg-slate-800 border border-slate-700 flex items-center justify-center font-bold text-blue-400">
                FA
              </div>
            </div>
          </div>
        </header>

        {/* Dashboard Grid */}
        <div className="flex-1 overflow-y-auto p-8 space-y-8">
          {/* Stats Grid */}
          <div className="grid grid-cols-4 gap-6">
            {[
              { label: 'Active Drivers', value: '12', icon: Users, color: 'text-blue-400', bg: 'bg-blue-400/10' },
              { label: 'Safety Score', value: '88%', icon: Shield, color: 'text-emerald-400', bg: 'bg-emerald-400/10' },
              { label: 'Alerts (24h)', value: '142', icon: AlertTriangle, color: 'text-amber-400', bg: 'bg-amber-400/10' },
              { label: 'Live Telemetry', value: '4.2k', icon: Activity, color: 'text-indigo-400', bg: 'bg-indigo-400/10' },
            ].map((stat, i) => (
              <div key={i} className="glass rounded-2xl p-6 border border-slate-800 hover:border-slate-700 transition-all group">
                <div className="flex justify-between items-start mb-4">
                  <div className={`${stat.bg} p-3 rounded-xl group-hover:scale-110 transition-transform`}>
                    <stat.icon className={`w-6 h-6 ${stat.color}`} />
                  </div>
                  <TrendingUp className="w-4 h-4 text-emerald-500" />
                </div>
                <p className="text-slate-400 text-sm font-medium mb-1">{stat.label}</p>
                <h3 className="text-2xl font-bold">{stat.value}</h3>
              </div>
            ))}
          </div>

          <div className="grid grid-cols-3 gap-8">
            {/* Live Map Mockup */}
            <div className="col-span-2 glass rounded-3xl p-6 border border-slate-800 min-h-[400px] flex flex-col">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-bold flex items-center gap-2">
                  <Eye className="w-5 h-5 text-rose-400" />
                  Violence Monitor Feed
                </h3>
                <div className="flex gap-2">
                  <span className="px-3 py-1 bg-rose-500/10 rounded-lg text-xs font-semibold text-rose-400 border border-rose-500/20 flex items-center gap-2">
                    <div className="w-1.5 h-1.5 rounded-full bg-rose-500 animate-pulse" />
                    LIVE AI ANALYSIS
                  </span>
                </div>
              </div>
              
              {/* Camera Feed */}
              <div className="flex-1 bg-slate-900 rounded-2xl border border-slate-800 relative overflow-hidden group">
                <img 
                  src="http://localhost:5000/video_feed" 
                  alt="Security Feed"
                  className="w-full h-full object-cover"
                  onError={(e) => {
                    (e.target as HTMLImageElement).src = 'https://images.unsplash.com/photo-1557597774-9d2739f85a76?auto=format&fit=crop&q=80&w=1000';
                    (e.target as HTMLImageElement).style.opacity = '0.3';
                  }}
                />
                
                {/* Overlay UI */}
                <div className="absolute top-4 right-4 flex gap-2">
                   <div className="bg-black/60 backdrop-blur-md px-3 py-1.5 rounded-lg border border-white/10 flex items-center gap-2">
                      <div className="w-2 h-2 rounded-full bg-emerald-500" />
                      <span className="text-[10px] font-bold text-white uppercase tracking-wider">CAM-01 ACTIVE</span>
                   </div>
                </div>

                <div className="absolute bottom-4 left-4 right-4 p-4 glass rounded-xl flex items-center justify-between">
                   <div className="flex items-center gap-4">
                      <div className="bg-blue-500/20 p-2 rounded-lg">
                        <Activity className="w-4 h-4 text-blue-400" />
                      </div>
                      <div>
                        <p className="text-[10px] font-bold text-slate-500 uppercase">Current Source</p>
                        <p className="text-xs font-semibold text-white">DroidCam Mobile Feed</p>
                      </div>
                   </div>
                   <div className="text-right">
                      <p className="text-[10px] font-bold text-slate-500 uppercase">Detection Status</p>
                      <p className="text-xs font-semibold text-emerald-400">Monitoring for Violence</p>
                   </div>
                </div>
              </div>
            </div>

            {/* Real-time Alert Feed */}
            <div className="glass rounded-3xl p-6 border border-slate-800 flex flex-col h-[500px]">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-bold flex items-center gap-2">
                  <AlertTriangle className="w-5 h-5 text-amber-400" />
                  Live Alerts
                </h3>
                <span className="bg-slate-800 px-2 py-1 rounded text-[10px] font-bold text-slate-400">REALTIME</span>
              </div>

              <div className="flex-1 overflow-y-auto pr-2 space-y-4 scrollbar-hide">
                <AnimatePresence initial={false}>
                  {alerts.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full text-center p-6 opacity-30">
                      <Wifi className="w-10 h-10 mb-4" />
                      <p className="text-sm">Listening for alerts from the fleet...</p>
                    </div>
                  ) : (
                    alerts.map((alert, i) => (
                      <motion.div
                        key={i}
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        className={`p-4 rounded-2xl border ${
                          alert.severity === 'CRITICAL' 
                          ? 'bg-rose-500/5 border-rose-500/20' 
                          : alert.severity === 'HIGH' 
                          ? 'bg-amber-500/5 border-amber-500/20' 
                          : 'bg-slate-800/30 border-slate-700/30'
                        }`}
                      >
                        <div className="flex justify-between items-start mb-2">
                          <span className={`text-[10px] font-bold px-2 py-0.5 rounded uppercase ${
                            alert.severity === 'CRITICAL' ? 'bg-rose-500 text-white' : 'bg-slate-700 text-slate-300'
                          }`}>
                            {alert.severity}
                          </span>
                          <span className="text-[10px] text-slate-500 font-mono">
                            {new Date(alert.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                          </span>
                        </div>
                        <p className="text-sm font-semibold text-slate-100">{alert.message}</p>
                        <div className="mt-2 flex items-center gap-2 text-[10px] text-slate-500 font-medium">
                          <MapPin className="w-3 h-3" />
                          <span>Vehicle ID: V-4921</span>
                        </div>
                      </motion.div>
                    ))
                  )}
                </AnimatePresence>
              </div>
            </div>
          </div>

          {/* Analytics Section */}
          <div className="glass rounded-3xl p-8 border border-slate-800">
            <div className="flex items-center justify-between mb-10">
              <div>
                <h3 className="text-xl font-bold mb-1">Fleet Analytics</h3>
                <p className="text-slate-400 text-sm">Performance metrics across the last 24 hours</p>
              </div>
              <div className="flex gap-4">
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded-full bg-blue-500" />
                  <span className="text-xs text-slate-400">Total Alerts</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded-full bg-emerald-500" />
                  <span className="text-xs text-slate-400">Safety Index</span>
                </div>
              </div>
            </div>

            <div className="h-[300px] w-full">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={analyticsData}>
                  <defs>
                    <linearGradient id="colorAlerts" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.1}/>
                      <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                    </linearGradient>
                    <linearGradient id="colorScore" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#10b981" stopOpacity={0.1}/>
                      <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" vertical={false} />
                  <XAxis dataKey="name" stroke="#64748b" fontSize={12} tickLine={false} axisLine={false} />
                  <YAxis stroke="#64748b" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip 
                    contentStyle={{ backgroundColor: '#0f172a', borderColor: '#1e293b', borderRadius: '12px' }}
                    itemStyle={{ fontSize: '12px', fontWeight: 'bold' }}
                  />
                  <Area type="monotone" dataKey="alerts" stroke="#3b82f6" strokeWidth={3} fillOpacity={1} fill="url(#colorAlerts)" />
                  <Area type="monotone" dataKey="score" stroke="#10b981" strokeWidth={3} fillOpacity={1} fill="url(#colorScore)" />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
