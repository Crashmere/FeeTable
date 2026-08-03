package com.example.feetable

import android.app.Application
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.example.feetable.data.AppDatabase
import com.example.feetable.data.repository.FeeRepository
import com.example.feetable.ui.editor.TableEditorScreen
import com.example.feetable.ui.export.ExportScreen
import com.example.feetable.ui.home.HomeScreen
import com.example.feetable.ui.theme.FeeTableTheme

class FeeTableApplication : Application() {
    val database by lazy { AppDatabase.getDatabase(this) }
    val repository by lazy {
        FeeRepository(
            database.feeTableDao(),
            database.feeRecordDao(),
            database.locationDao(),
            database.tagDao()
        )
    }
}

@Composable
fun FeeTableApp() {
    FeeTableTheme {
        Surface(
            modifier = Modifier.fillMaxSize(),
            color = MaterialTheme.colorScheme.background
        ) {
            val navController = rememberNavController()

            NavHost(navController = navController, startDestination = "home") {
                composable("home") {
                    HomeScreen(
                        onTableClick = { tableId ->
                            navController.navigate("editor/$tableId")
                        }
                    )
                }
                composable(
                    "editor/{tableId}",
                    arguments = listOf(navArgument("tableId") { type = NavType.IntType })
                ) { backStackEntry ->
                    val tableId = backStackEntry.arguments?.getInt("tableId") ?: return@composable
                    TableEditorScreen(
                        tableId = tableId,
                        onBack = { navController.popBackStack() },
                        onExport = { navController.navigate("export/$tableId") },
                        onExportByTag = { tag ->
                            navController.navigate("export/$tableId/$tag")
                        }
                    )
                }
                composable(
                    "export/{tableId}",
                    arguments = listOf(navArgument("tableId") { type = NavType.IntType })
                ) { backStackEntry ->
                    val tableId = backStackEntry.arguments?.getInt("tableId") ?: return@composable
                    ExportScreen(
                        tableId = tableId,
                        filterTag = null,
                        onBack = { navController.popBackStack() }
                    )
                }
                composable(
                    "export/{tableId}/{filterTag}",
                    arguments = listOf(
                        navArgument("tableId") { type = NavType.IntType },
                        navArgument("filterTag") { type = NavType.StringType }
                    )
                ) { backStackEntry ->
                    val tableId = backStackEntry.arguments?.getInt("tableId") ?: return@composable
                    val filterTag = backStackEntry.arguments?.getString("filterTag")
                    ExportScreen(
                        tableId = tableId,
                        filterTag = filterTag,
                        onBack = { navController.popBackStack() }
                    )
                }
            }
        }
    }
}
