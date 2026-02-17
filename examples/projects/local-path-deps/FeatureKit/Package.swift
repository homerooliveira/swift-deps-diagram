// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "FeatureKit",
    products: [
        .library(name: "FeatureKit", targets: ["FeatureKit"]),
    ],
    dependencies: [
        .package(name: "CoreKit", path: "../CoreKit"),
    ],
    targets: [
        .target(
            name: "FeatureKit",
            dependencies: ["CoreKit"]
        ),
    ]
)
