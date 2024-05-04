import { useEffect, useState } from "react";
import { Button, Pressable, StyleSheet, Text, View, PermissionsAndroid } from "react-native";
import { useVoiceRecognition } from "../hooks/useVoiceRecognition";
import * as FileSystem from "expo-file-system"
import { Audio } from "expo-av";
import { writeAudioToFile } from "../utils/writeAudioToFile";
import { playFromPath } from "../utils/playFromPath";
import { fetchAudio } from "../utils/fetchAudio";
import { fetchConversation } from "../utils/fetchConversation";
import Geolocation from 'react-native-geolocation-service';

Audio.setAudioModeAsync({
  allowsRecordingIOS: false,
  staysActiveInBackground: false,
  playsInSilentModeIOS: true,
  shouldDuckAndroid: true,
  playThroughEarpieceAndroid: false,
});

const requestLocationPermission = async () => {
  try {
    const granted = await PermissionsAndroid.request(
      PermissionsAndroid.PERMISSIONS.ACCESS_FINE_LOCATION,
      {
        title: 'Geolocation Permission',
        message: 'Can we access your location?',
        buttonNeutral: 'Ask Me Later',
        buttonNegative: 'Cancel',
        buttonPositive: 'OK',
      },
    );

    return granted === 'granted';
  } catch (err) {
    return false;
  }
};

let intervalHandle: any = null;

export default function MainScreen() {
  const [borderColor, setBorderColor] = useState<"lightgray" | "lightgreen">("lightgray");
  const {state, startRecognizing, stopRecognizing, destroyRecognizer} = useVoiceRecognition();
  const [urlPath, setUrlPath] = useState("");
  const [convId, setConvId] = useState("");
  const [location, setLocation] = useState({
    isValid: false,
    latitude: 0,
    longitude: 0,
  });

  const getLocation = () => {
    const result = requestLocationPermission();
    result.then(res => {
      if (res) {
        Geolocation.getCurrentPosition(
          position => {
            console.log(position.coords);
            setLocation({
              isValid: true,
              latitude: position.coords.latitude,
              longitude: position.coords.longitude,
            });
          },
          error => {
            // See error code charts below.
            console.log(error.code, error.message);
            setLocation({
              isValid: false,
              latitude: 0,
              longitude: 0,
            });
          },
          {enableHighAccuracy: true, timeout: 15000, maximumAge: 10000},
        );
      }
    });
  };

  useEffect(() => {
    if (intervalHandle === null) {
      getLocation();
      intervalHandle = setInterval(getLocation, 30000);
    }
  })

  const listFiles = async () => {
    try {
      const result = await FileSystem.readDirectoryAsync(FileSystem.documentDirectory!);
      if (result.length > 0) {
        const fileName = result[0];
        const path = FileSystem.documentDirectory + fileName;
        setUrlPath(path);
      }
    } catch(e) {
      console.log(e);
    }
  }

  const handleSubmit = async () => {
    if (!state.results[0]) {
      return;
    }

    console.log(location)

    try {
      if (convId) {
        talk(state.results[0], convId);
      } else {
        const convResponse = await fetchConversation()
        if (convResponse?.conversation?.id) {
          setConvId(convResponse.conversation.id);
          talk(state.results[0], convResponse.conversation.id);
        }
      }
    } catch(err) {
      console.log(err);
    }
  }

  const talk = async (prompt: string, convId: string) => {
    console.log(convId, prompt);

    let coords = null;
    if (location.isValid) {
      coords = location;
    }
    const response: any = await fetchAudio(prompt, convId, coords);
      const reader = new FileReader();
      reader.onload = async (e) => {
        if (e.target && typeof e.target.result === "string") {
          const audioData = e.target.result.split(",")[1];
          const path = await writeAudioToFile(audioData);
          setUrlPath(path)
          await playFromPath(path);
          destroyRecognizer();
        }
      };
      reader.readAsDataURL(response);
  }

  return (
    <View style={styles.container}>
      <Text style={{ fontSize: 32, fontWeight: "bold", marginBottom: 30 }}>Talk to me</Text>
      <Text>Press and hold this button to speak</Text>
      <Pressable 
        style={[styles.pressable, {borderColor: borderColor}]}
        onPressIn={() => {
            setBorderColor("lightgreen");
            startRecognizing();
        }}
        onPressOut={() => {
            setBorderColor("lightgray");
            stopRecognizing();
            handleSubmit();
        }}
      >
        <Text>Hold to speak</Text>
      </Pressable>
      <Text style={styles.message}>Your message: {state.results.length > 0 ? state.results[0] : ""}</Text>
      <Button title="Replay last message" onPress={async () => {
        await playFromPath(urlPath);
      }} />



      {/* <Text>{JSON.stringify(state, null, 2)}</Text> */}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 20,
  },
  message: {
    textAlign: 'center',
    color: "#0000ff",
    marginVertical: 10,
    fontSize: 17,
    fontWeight: "bold",
  },
  pressable: {
    width: "90%",
    padding: 30,
    gap: 10,
    borderWidth: 3,
    alignItems: "center",
    borderRadius: 10,
  },
});

  