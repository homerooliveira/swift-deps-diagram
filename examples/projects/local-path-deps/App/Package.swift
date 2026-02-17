// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "ExampleApp",
    products: [
        .executable(name: "ExampleApp", targets: ["ExampleApp"]),
    ],
    dependencies: [
        .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.8.0"),
        .package(name: "FeatureKit", path: "../FeatureKit"),
    ],
    targets: [
        .executableTarget(
            name: "ExampleApp",
            dependencies: [
                .product(name: "Alamofire", package: "alamofire"),
                .product(name: "FeatureKit", package: "FeatureKit"),
            ]
        ),
    ]
)
