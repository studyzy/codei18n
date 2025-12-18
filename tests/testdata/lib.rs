//! DAO Fork related constants from [EIP-779](https://eips.ethereum.org/EIPS/eip-779).
//! It happened on Ethereum block 1_920_000

//! Inner doc comment for module

/// Outer doc comment for function
fn main() {
    // Line comment
    /* Block
    comment */
    // good
    // luck
}

/// A mapping of precompile contracts that can be either static (builtin) or dynamic.
///
/// This is an optimization that allows us to keep using the static precompiles
/// until we need to modify them, at which point we convert to the dynamic representation.
struct PrecompileContract {
    address: u64,
    name: String,
}