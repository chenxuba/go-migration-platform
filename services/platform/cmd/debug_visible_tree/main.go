package main

import (
    "context"
    "database/sql"
    "fmt"
    "sort"
    "strings"

    _ "github.com/go-sql-driver/mysql"
    "go-migration-platform/pkg/config"
    "go-migration-platform/services/platform/internal/model"
    "go-migration-platform/services/platform/internal/repository"
)

func main() {
    cfg := config.Load("platform-service", "8082")
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
    db, err := sql.Open("mysql", dsn)
    if err != nil { panic(err) }
    defer db.Close()

    repo, err := repository.New(db)
    if err != nil { panic(err) }

    tree, err := repo.ListModuleMenuTree(context.Background(), 1)
    if err != nil { panic(err) }

    var total int
    var pageUse int
    var paths []string
    var childCounts []string
    for _, group := range tree {
        for _, child := range group.Children {
            childCounts = append(childCounts, fmt.Sprintf("%s > %s = %d", group.MenuName, child.MenuName, len(child.Children)))
            for _, leaf := range child.Children {
                total++
                path := fmt.Sprintf("%s > %s > %s", group.MenuName, child.MenuName, leaf.MenuName)
                paths = append(paths, path)
                if strings.TrimSpace(leaf.MenuName) == "页面功能访问" {
                    pageUse++
                }
            }
        }
    }
    sort.Strings(paths)
    sort.Strings(childCounts)
    fmt.Printf("TOTAL=%d\n", total)
    fmt.Printf("PAGE_USE=%d\n", pageUse)
    fmt.Println("--- CHILD COUNTS ---")
    for _, line := range childCounts {
        fmt.Println(line)
    }
    fmt.Println("--- PAGE_USE PATHS ---")
    for _, path := range paths {
        if strings.HasSuffix(path, "> 页面功能访问") {
            fmt.Println(path)
        }
    }
}

var _ model.ModuleMenu
