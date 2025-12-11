

import cv2
import mediapipe as mp
import numpy as np
from typing import Tuple, List, Optional


class InterviewAssistant:
    """AI assistant that monitors eye contact and posture during interviews."""
    
    def __init__(self):

        self.mp_face_mesh = mp.solutions.face_mesh
        self.mp_pose = mp.solutions.pose
        self.mp_drawing = mp.solutions.drawing_utils
        self.mp_drawing_styles = mp.solutions.drawing_styles
        

        self.face_mesh = self.mp_face_mesh.FaceMesh(
            max_num_faces=1,
            refine_landmarks=True,
            min_detection_confidence=0.5,
            min_tracking_confidence=0.5
        )
        
        
        self.pose = self.mp_pose.Pose(
            min_detection_confidence=0.5,
            min_tracking_confidence=0.5
        )
        

        self.LEFT_EYE_INDICES = [33, 133, 159, 145, 158, 153]
        self.RIGHT_EYE_INDICES = [362, 263, 386, 374, 387, 380]
        
        
        self.LEFT_EYE_CENTER = [33, 133, 157, 158, 159, 160, 161, 246]
        self.RIGHT_EYE_CENTER = [362, 263, 388, 387, 386, 385, 384, 398]
        
       
        self.posture_history = []
        self.eye_contact_history = []
        
    def calculate_eye_aspect_ratio(self, landmarks, eye_indices: List[int], image_shape: Tuple[int, int]) -> float:
        """Calculate Eye Aspect Ratio (EAR) to detect eye openness.
        Uses 6 points: [p1, p2, p3, p4, p5, p6]
        EAR = (|p2-p6| + |p3-p5|) / (2 * |p1-p4|)
        """
        eye_points = []
        for idx in eye_indices:
            landmark = landmarks.landmark[idx]
            x = landmark.x * image_shape[1]
            y = landmark.y * image_shape[0]
            eye_points.append([x, y])
        
        eye_points = np.array(eye_points)
        

        vertical_1 = np.linalg.norm(eye_points[1] - eye_points[5])  # p2 to p6
        vertical_2 = np.linalg.norm(eye_points[2] - eye_points[4])  # p3 to p5
        
        horizontal = np.linalg.norm(eye_points[0] - eye_points[3])   # p1 to p4
        
        
        if horizontal == 0:
            return 0.0
        ear = (vertical_1 + vertical_2) / (2.0 * horizontal)
        return ear
    
    def calculate_eye_direction(self, landmarks, eye_center_indices: List[int], image_shape: Tuple[int, int]) -> Tuple[float, float]:
        """Calculate eye direction (gaze direction) relative to camera."""
        eye_center_points = []
        
        for idx in eye_center_indices:
            if idx < len(landmarks.landmark):
                landmark = landmarks.landmark[idx]
                x = landmark.x * image_shape[1]
                y = landmark.y * image_shape[0]
                eye_center_points.append([x, y])
        
        if not eye_center_points:
            return 0.0, 0.0
        
        eye_center = np.mean(eye_center_points, axis=0)
        image_center = np.array([image_shape[1] / 2, image_shape[0] / 2])
        
       
        offset = eye_center - image_center
      
        normalized_offset = offset / image_center
        
        return normalized_offset[0], normalized_offset[1]  # x, y offsets
    
    def analyze_posture(self, pose_landmarks, image_shape: Tuple[int, int]) -> dict:
        """Analyze posture using pose landmarks."""
        if not pose_landmarks:
            return None
        
        landmarks = pose_landmarks.landmark
        
       
        left_shoulder = landmarks[self.mp_pose.PoseLandmark.LEFT_SHOULDER]
        right_shoulder = landmarks[self.mp_pose.PoseLandmark.RIGHT_SHOULDER]
        
       ts
        left_hip = landmarks[self.mp_pose.PoseLandmark.LEFT_HIP]
        right_hip = landmarks[self.mp_pose.PoseLandmark.RIGHT_HIP]
        
        
        shoulder_y_diff = abs(left_shoulder.y - right_shoulder.y)
        shoulder_alignment = "aligned" if shoulder_y_diff < 0.05 else "misaligned"
        
        
        shoulder_center_y = (left_shoulder.y + right_shoulder.y) / 2
        hip_center_y = (left_hip.y + right_hip.y) / 2
        
        
        forward_lean = shoulder_center_y - hip_center_y
        
        
        posture_quality = "good"
        issues = []
        
        if shoulder_y_diff > 0.1:
            posture_quality = "poor"
            issues.append("Shoulders not level")
        
        if forward_lean > 0.15:
            posture_quality = "poor"
            issues.append("Leaning too far forward")
        elif forward_lean < -0.1:
            posture_quality = "poor"
            issues.append("Leaning too far back")
        
        return {
            "quality": posture_quality,
            "shoulder_alignment": shoulder_alignment,
            "forward_lean": forward_lean,
            "issues": issues
        }
    
    def process_frame(self, frame) -> Tuple[np.ndarray, dict]:
        """Process a single frame and return annotated frame with analysis."""
        rgb_frame = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)
        results_face = self.face_mesh.process(rgb_frame)
        results_pose = self.pose.process(rgb_frame)
        
        analysis = {
            "eye_contact": None,
            "eye_open": None,
            "posture": None
        }
        
        
        if results_pose.pose_landmarks:
            self.mp_drawing.draw_landmarks(
                frame,
                results_pose.pose_landmarks,
                self.mp_pose.POSE_CONNECTIONS,
                landmark_drawing_spec=self.mp_drawing_styles.get_default_pose_landmarks_style()
            )
            
            
            analysis["posture"] = self.analyze_posture(
                results_pose.pose_landmarks,
                frame.shape
            )
        
        
        if results_face.multi_face_landmarks:
            for face_landmarks in results_face.multi_face_landmarks:
               
                self.mp_drawing.draw_landmarks(
                    frame,
                    face_landmarks,
                    self.mp_face_mesh.FACEMESH_CONTOURS,
                    None,
                    self.mp_drawing_styles.get_default_face_mesh_contours_style()
                )
                
                
                left_ear = self.calculate_eye_aspect_ratio(
                    face_landmarks,
                    self.LEFT_EYE_INDICES,
                    frame.shape
                )
                right_ear = self.calculate_eye_aspect_ratio(
                    face_landmarks,
                    self.RIGHT_EYE_INDICES,
                    frame.shape
                )
                
                avg_ear = (left_ear + right_ear) / 2.0
                eye_open = avg_ear > 0.25  # Threshold for eye open/closed
                analysis["eye_open"] = eye_open
                
                
                left_eye_dir = self.calculate_eye_direction(
                    face_landmarks,
                    self.LEFT_EYE_CENTER,
                    frame.shape
                )
                right_eye_dir = self.calculate_eye_direction(
                    face_landmarks,
                    self.RIGHT_EYE_CENTER,
                    frame.shape
                )
                
                avg_gaze_x = (left_eye_dir[0] + right_eye_dir[0]) / 2
                avg_gaze_y = (left_eye_dir[1] + right_eye_dir[1]) / 2
                
                # Determine eye contact (looking at camera)
                # If gaze is close to center (within threshold)
                eye_contact = abs(avg_gaze_x) < 0.15 and abs(avg_gaze_y) < 0.15
                analysis["eye_contact"] = eye_contact
                
                # Draw eye contact indicator
                if eye_contact:
                    cv2.circle(frame, (50, 50), 20, (0, 255, 0), -1)
                    cv2.putText(frame, "Eye Contact", (80, 55), 
                               cv2.FONT_HERSHEY_SIMPLEX, 0.7, (0, 255, 0), 2)
                else:
                    cv2.circle(frame, (50, 50), 20, (0, 0, 255), -1)
                    cv2.putText(frame, "No Eye Contact", (80, 55), 
                               cv2.FONT_HERSHEY_SIMPLEX, 0.7, (0, 0, 255), 2)
        
        # Display posture information
        if analysis["posture"]:
            posture = analysis["posture"]
            y_offset = 100
            cv2.putText(frame, f"Posture: {posture['quality'].upper()}", 
                       (10, y_offset), cv2.FONT_HERSHEY_SIMPLEX, 0.7, 
                       (0, 255, 0) if posture['quality'] == 'good' else (0, 0, 255), 2)
            
            if posture['issues']:
                for i, issue in enumerate(posture['issues']):
                    cv2.putText(frame, f"- {issue}", (10, y_offset + 30 + i * 25),
                               cv2.FONT_HERSHEY_SIMPLEX, 0.5, (0, 0, 255), 1)
        
        # Display eye status
        if analysis["eye_open"] is not None:
            eye_status = "Eyes Open" if analysis["eye_open"] else "Eyes Closed"
            cv2.putText(frame, eye_status, (10, frame.shape[0] - 30),
                       cv2.FONT_HERSHEY_SIMPLEX, 0.7, (255, 255, 255), 2)
        
        return frame, analysis
    
    def run(self):
        """Run the interview assistant with webcam."""
        cap = cv2.VideoCapture(0)
        
        if not cap.isOpened():
            print("Error: Could not open webcam")
            return
        
        print("Interview Assistant started. Press 'q' to quit.")
        
        while True:
            ret, frame = cap.read()
            if not ret:
                break
            
            frame, analysis = self.process_frame(frame)
            
            # Display frame
            cv2.imshow('Interview AI Assistant', frame)
            
            # Print analysis to console
            if analysis["eye_contact"] is not None:
                print(f"Eye Contact: {analysis['eye_contact']}, "
                      f"Eyes Open: {analysis['eye_open']}")
            if analysis["posture"]:
                print(f"Posture: {analysis['posture']['quality']}")
            
            # Exit on 'q' key
            if cv2.waitKey(1) & 0xFF == ord('q'):
                break
        
        cap.release()
        cv2.destroyAllWindows()


if __name__ == "__main__":
    assistant = InterviewAssistant()
    assistant.run()
