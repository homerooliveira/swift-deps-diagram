// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "HelloSPM",
    products: [
        .executable(name: "HelloSPM", targets: ["HelloSPM"]),
    ],
    dependencies: [
        .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.8.0"),
    ],
    targets: [
        .executableTarget(
            name: "HelloSPM",
            dependencies: [
                .product(name: "Alamofire", package: "alamofire"),
            ]
        ),
    ]
)
