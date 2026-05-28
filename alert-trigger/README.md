# Alert Trigger App

A React Native Expo application to trigger various AI safety alerts on the ThirdEye backend.

## Features
- **ADAS Alerts**: Pedestrian, Theft, Collision, Recording.
- **DMS Alerts**: Drowsiness, Distraction, Mobile Usage, Helmet Detection, Camera Blocked, Smoking.
- **Vehicle Dynamics**: Speeding, Harsh Braking, Harsh Acceleration, Sharp Turning.
- **Driver Management**: Login and Face ID Registration simulation.

## Backend
- **Base URL**: `http://139.59.73.32:5023`
- The app sends GET requests to the respective endpoints to trigger events.

## Getting Started

1. Install dependencies:
   ```bash
   npm install
   ```

2. Start the Expo server:
   ```bash
   npx expo start
   ```

3. Open on your device:
   - Use the Expo Go app on iOS or Android.
   - Scan the QR code displayed in the terminal.

## UI Design
- Premium Dark Mode
- Glassmorphism effects
- Haptic feedback on interactions
- Categorized grid layout for quick access
