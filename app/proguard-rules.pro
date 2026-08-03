# Disable obfuscation (renaming) - POI uses reflection/ServiceLoader patterns
# that break when class names change. Shrinking (removing unused code) still works.
-dontobfuscate

# Room
-keep class * extends androidx.room.RoomDatabase { *; }
-keep @androidx.room.Entity class * { *; }
-keep @androidx.room.Dao interface * { *; }

# Apache POI
-keep class org.apache.poi.xssf.usermodel.XSSFWorkbook { *; }
-keep class org.apache.poi.xssf.usermodel.XSSFSheet { *; }
-keep class org.apache.poi.xssf.usermodel.XSSFRow { *; }
-keep class org.apache.poi.xssf.usermodel.XSSFCell { *; }
-keep class org.apache.poi.xssf.usermodel.XSSFCellStyle { *; }
-keep class org.apache.poi.xssf.usermodel.XSSFFont { *; }
-keep class org.apache.poi.xssf.usermodel.XSSFDataFormat { *; }
-keep class org.apache.poi.ss.usermodel.** { *; }
-keep class org.apache.poi.ss.util.** { *; }
-keep class org.apache.poi.** { *; }
-keep class org.apache.xmlbeans.** { *; }
-keep class org.openxmlformats.** { *; }
-dontwarn org.apache.poi.**
-dontwarn org.apache.xmlbeans.**
-dontwarn org.openxmlformats.**
-dontwarn org.etsi.**
-dontwarn org.w3.**
-dontwarn com.microsoft.**
-dontwarn org.apache.batik.**
-dontwarn org.apache.pdfbox.**
-dontwarn org.apache.commons.compress.**
-dontwarn org.apache.commons.codec.**
-dontwarn org.apache.commons.io.**
-dontwarn org.apache.commons.math3.**
-dontwarn javax.xml.stream.**
-dontwarn org.apache.logging.**
-dontwarn java.awt.**
-dontwarn javax.swing.**
-dontwarn com.graphbuilder.**

# ViewModels (instantiated via reflection)
-keep class com.example.feetable.ui.** { *; }
-keep class com.example.feetable.FeeTableApplication { *; }

# Compose - keep animation core to prevent R8 stripping keyframe methods
-keep class androidx.compose.animation.** { *; }

# Kotlin Coroutines
-keepnames class kotlinx.coroutines.internal.MainDispatcherFactory {}
-keepnames class kotlinx.coroutines.CoroutineExceptionHandler {}

# Kotlin serialization & reflection
-keepattributes *Annotation*
-keepattributes Signature
-keepattributes InnerClasses
