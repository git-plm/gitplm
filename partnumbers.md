# Internal Part number guidelines

<!-- vim-markdown-toc GFM -->

- [Why have internal part numbers (IPNs)](#why-have-internal-part-numbers-ipns)
- [IPN Requirements](#ipn-requirements)
- [Who uses IPNs and how are they used?](#who-uses-ipns-and-how-are-they-used)
- [Suggested part number format](#suggested-part-number-format)
  - [Encoding Product Version and Variations](#encoding-product-version-and-variations)
  - [Three letter category code](#three-letter-category-code)
  - [Three number category code](#three-number-category-code)
  - [Two letter category code](#two-letter-category-code)
  - [Why use the same format for IPN and external model number?](#why-use-the-same-format-for-ipn-and-external-model-number)
  - [Examples](#examples)
    - [Resistor part numbers](#resistor-part-numbers)
    - [Capacitor part numbers](#capacitor-part-numbers)
- [Implementation](#implementation)
- [Structured or Unstructured?](#structured-or-unstructured)
- [Reference](#reference)

<!-- vim-markdown-toc -->

## Why have internal part numbers (IPNs)

An Internal Part Numbers (IPNs) (also call stock codes, unique identifier codes,
etc) is a number that identifies all parts, components, software, and
documentation required to manufacture and support a product. It can also
identify finished products.

Companies that manufacture products typically use IPNs for the following
reasons:

- they need to be able to identify and organize parts/products they use and
  manufacture
- it is useful to assign IPNs to purchased parts as well so you can easily
  change sources, have multiple sources, etc.
- any time physical materials are handled, they need to be labeled in a way that
  is consistent so that it is easy to recognize and process by humans.

Software, hardware, documentation, and about everything else that is used to
make products can and does change over time. However, an IPN is like your given
name at birth -- it's hard to change later -- especially the format. Thus, it is
good to think through this -- what are your goals, who uses IPNs, how are they
used, how is material handled, etc. IPNs seem trivial -- just a few
characters/digits -- why does it matter? What is this in relation to the
complexity of the design, software, etc? But it is the most basic, simple things
that matter the most, especially if you want to scale.

As Phil Karlton said:

> There are only two hard things in Computer Science: cache invalidation and
> naming things.

What are names in a program for? Humans. Compilers certainly don't need good
names. **IPNs likewise are for humans**; otherwise, we'd just use a UUID.
Perhaps it would be more correct to say Internal Part **Name** instead of
Number. Naming is hard -- don't expect IPNs to be different. Take your time to
think it through and get it right. Don't limit yourself to what a certain tool
can support, or what is best for one department -- make sure it's great for the
entire organization. An optimized IPN strategy can **improve your organization's
efficiencies**, and **reduce mistakes.**

## IPN Requirements

- short enough to remember for commonly used parts
- easy to identify, compare, and process by humans
- easy to recognize as an IPN and differentiate from other numbers
- structured to minimize errors/mistakes
- handles part/product **variations**
- handles part/product **versions**

## Who uses IPNs and how are they used?

- **Engineering**: create part numbers, and requirements for a component. Also
  create BOMs that specify how purchased parts, manufactured parts, software,
  documentation, etc are combined into a product.
- **Purchasing**: order components/parts
- **Manufacturing**: material planning and execute the process of combining
  parts into products
- **Marketing**: organize products/service parts for sale and provide
  identifiers for customers to order products.
- **Sales/Shipping**: process orders and ship products to customers

## Suggested part number format

Basic format: CCC-NNN-VVVV

- CCC: one to three letters or numbers to identify major category (RES, CAP,
  DIO, E (electrical), M (mechanical), etc).
- NNN: incrementing sequential number for each part. This gives this format
  flexibility.
- VVVV: use to code variations of similar parts typically with the **same
  datasheet or family** (resistance, capacitance, regulator voltage, IC package,
  screw type, etc.). VVVV is also used to encode version and variations for
  manufactured parts or assemblies.

This PN format can be used to track a wide range of items:

- components you order
- software/firmware that gets loaded in a product
- manufacturing documentation
- parts/products you manufacture
- model numbers that get listed on your website

The dash is the suggested separator in IPNs over a period or underscore for the
following reasons:

- BOMs are often imported as spreadsheets, and a period may be interpreted as a
  decimal point in a number
- Dashs are much easier to see than underscores

### Encoding Product Version and Variations

Manufactured products often are sold as a family of related products. For
example, you might manufacture a gateway with different RAM and Flash
configurations. It is useful to group these product IPNs together as
_variations_ of a product. Products also undergo changes over time (_versions_).

The 4 digit variation field can be used to encode both of these. The first two
digits can be used to encode the variation, and the second two digits the
version:

| IPN          | RAM | Flash | Version |
| ------------ | --- | ----- | ------- |
| GTW-012-0001 | 64  | 64    | 1       |
| GTW-012-0101 | 128 | 64    | 1       |
| GTW-012-0201 | 128 | 128   | 1       |
| GTW-012-0002 | 64  | 64    | 2       |
| GTW-012-0102 | 128 | 64    | 2       |
| GTW-012-0202 | 128 | 128   | 2       |

This gives us 100 variations, and 100 versions. If a product exceeds these
limits, then celebrate that you are wildly successful!!!! Your product has been
in production for a long time or has enough volume to justify a large number of
variations.

### Three letter category code

The following is just an example -- will likely need tweaked for different
industries.

The number of categories should be kept reasonably small to minimize the
difficulty of assigning new part numbers.

The following CCC groups are suggested for electrical parts:

| Code | Description                    |
| ---- | ------------------------------ |
| ANA  | op-amps, comparators, A/D, D/A |
| CAP  | capacitors                     |
| CON  | connectors                     |
| DIO  | diodes                         |
| IND  | inductors                      |
| ICS  | integrated circuits, mcus, etc |
| OSC  | oscillators                    |
| PWR  | relays, etc                    |
| REG  | regulators                     |
| RES  | resistors                      |
| SWI  | switch                         |
| TRA  | transistors, FETs              |
| TXT  | test points/pads               |
| XTL  | crystals                       |

Each group of CCC parts is placed in its own schematic symbol library with the
same name. Keeping the CCC and CAD library names the same introduces
consistency, which brings efficiency.

The following CCC groups are suggested for other parts (preliminary):

| Code | Description                                                                                                                                          |
| ---- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| FST  | fasteners (screws, bolts, nuts, etc)                                                                                                                 |
| CBL  | cables, etc                                                                                                                                          |
| ENC  | enclosures                                                                                                                                           |
| PKG  | packaging                                                                                                                                            |
| OPT  | Optics: Windows, Lens, light pipes, etc.                                                                                                             |
| FLD  | Fluids: Lubricants, oil, valve, check, divertor, reducer, tubes, pipes, hoses, seals, gaskets, sealants, diaphragms, bellows, pistons, cylinders     |
| MRK  | Markings: Coatings, labels, Pain, Dye, Ink, etc                                                                                                      |
| DRV  | Drive: Bearings/ Bushings, Gears and sprockets, Chains, Rollers, Motors, actuators                                                                   |
| STC  | Structural: connection hardware, lever arms, springs, beams, bars, plates, guide rods, ways, saddles, clamps, brackets, flanges, standoffs, castings |
| TMP  | Heat exchangers, sinks                                                                                                                               |

The following CCC groups are suggested for things you produce:

| Code | Description                                                                                                                                                                                                                                      |
| ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| PCS  | Printed Circuit Schematic.                                                                                                                                                                                                                       |
| PCA  | Printed Circuit Assembly. The version is incremented any time the BOM for the assembly changes.                                                                                                                                                  |
| PCB  | Printed Circuit board. This category identifies the bare PCB board.                                                                                                                                                                              |
| ASY  | Assembly (can be mechanical or top level subassembly -- typically represented by BOM and documentation). Again, the variation is incremented any time a BOM line item changes. You can also use product specific prefixes such as GTW (gateway). |
| DOC  | standalone documents                                                                                                                                                                                                                             |
| DFW  | data -- firmware to be loaded on MCUs, etc                                                                                                                                                                                                       |
| DSW  | data -- software (images for embedded Linux systems, applications, programming utilities, etc)                                                                                                                                                   |
| DCL  | data -- calibration data for a design                                                                                                                                                                                                            |
| FIX  | manufacturing fixtures                                                                                                                                                                                                                           |

Conventions can be used such that the PCS, PCA, and PCB NNN are a matched set:

| IPN          | Description                                    | Version |
| ------------ | ---------------------------------------------- | ------- |
| PCA-055-0002 | Gateway with RS485 support PCB assembly BOM    | 2       |
| PCA-055-0102 | Gateway with CAN support PCB assembly BOM      | 2       |
| PCB-055-0005 | Bare PCB used in above assemblies              | 5       |
| PCS-055-0006 | Schematic documentation for above PCB/assembly | 6       |

In the above, the common `055` ties all the IPNs together. We can quickly find
the schematic, bare PCB, or BOM if we know one of the IPNs -- whether it's a
file on disk, paper printout in the lab, documentation in the factory, field
service kit, etc.

Some additional guidelines:

- Every part number has the same number of characters in it (3-3-4). This makes
  sorting/comparison/entry simpler with less chance of error.
- Character set is restricted to capital letters, digits, and hyphen.
- Avoid punctuation characters such as %, !, (, ., etc.

### Three number category code

The following
[was suggested by a user](https://forum.kicad.info/t/internal-house-part-number-formats/34958/12?u=cliff_brake)
on the KiCad form.

The part category is a 3-position numeric field which designates the _type_ of
item. A summary of part categories is shown in Table 3 with the full list in
Appendix A.

| **Category** | **Description**                                   |
| ------------ | ------------------------------------------------- |
| **0xx-**     | Not used                                          |
| **1xx-**     | Assemblies                                        |
| **2xx-**     | Kits, Cables, Packages, PCB Fab, Mechanical Parts |
| **3xx-**     | Optical Components                                |
| **4xx-**     | Active Electrical Components                      |
| **5xx-**     | Passive Electrical Components                     |
| **6xx-**     | Standard Hardware Components                      |
| **7xx-**     | Misc. As Required Items                           |
| **8xx-**     | Software                                          |
| **9xx-**     | Documentation & Test                              |

### Two letter category code

| Code | Description   |
| ---- | ------------- |
| Ex   | Electrical    |
| Mx   | Mechanical    |
| Sx   | Software      |
| Px   | Prodcuct      |
| Dx   | Documentation |

x could be expanded to a number of sub categories -- perhaps sequentially
assigned.

### Why use the same format for IPN and external model number?

Having a consistent format between IPN and external model number has several
benefits:

- both should be easy for humans to recognize and process, so why not use the
  same format for both.
- both need to be stocked, warehoused, handled, etc
- customers may need to order service parts that are also used in manufacturing.
  If the internal and external numbers are the same, handling/stocking these
  parts is simpler.

### Examples

With resistors, capacitors, and connectors, we encode the value and pin count in
the variation:

- 1K 0805 1%: RES-002-1001
- 3.3K 0805 1%: RES-002-3301
- 2.2K 0603 5%: RES-003-2201 (note we bumped NNN to 003, because different
  package size)
- 10.3K high power 0603: RES-004-1032 (different vendor/datasheet than RES-002,
  so we bump NNN)
- 2x10, 0.1 in header: CON-000-0020
- 2x12, 0.1 in header: CON-000-0024
- 1x10, 0.1 in header: CON-001-0010
- 1x20, 0.1 in header: CON-001-0020

With most ICs we simply enumerate all the variations of a particular IC in a
sequentially incrementing variation (we don't try to encode information)

- LM78xx SOT223 5V: REG-089-0000
- LM78xx DIP 5V: REG-089-0001
- LM78xx SOT223 3.3V: REG-089-0002
- LM78xx DIP 3.3V: REG-089-0003
- 3.3v switching reg, SSOP8: REG-002-0000
- 3.3v switching reg, S08: REG-002-0001
- STM32H7 in 44 pin package, 1M flash: MCU-001-0000
- STM32H7 in 44 pin package, 2M flash: MCU-001-0001
- STM32H7 in 208 pin package, 1M flash: MCU-001-0002
- STM32H7 in 208 pin package, 2M flash: MCU-001-0003
- STM32F3 in 44 pin package: MCU-002-0000 (note different base part, so bump
  NNN)

Many parts will not have any variations:

- 2N4401 DIODE: DIO-000-0000 (no variation information, that is fine)
- 2N2222 transistor: TRA-000-0000 (again, no variation info)

The variation section is only used in cases where a part with a single datasheet
has multiple variations. Variations are generally used to encode one parameter
with the most different variations -- for instance resistance with resistors. A
single datasheet may include 0603, 0805, and 1206 options, but take out separate
NNN part numbers for different package sizes because with resistors, it makes
the most sense to encode the resistance in the variation (because there are lots
of resistance values), not the package size. There are relatively few package
sizes for resistors so it makes sense to take out new NNN numbers for different
packages. However, for voltage regulators, it may make sense to encode both the
regulated voltage and the package in the variation, because there is a
relatively small number of combinations.

Generally we don't need to create house part numbers for every part variation --
only the ones we use. Resistors/caps may be an exception where we simply create
the entire series in the partmaster because it is easiest to just do once.

#### Resistor part numbers

Most resistor variations (at least 1%) are encoded using the E96 4-digit
industry standard. Examples:

- 2500 = 250 x 100 = 250 x 1 = 250 Ω (This is only and only 250Ω not 2500 Ω)
- 1000 = 100 x 100 = 100x 1 = 100 Ω
- 7201 = 720 x 101 = 720 x 10 = 7200 Ω or 7.2kΩ
- 1001 = 100 × 101 =100 x 10 = 1000 Ω or 1kΩ
- 1004 = 100 × 104 =100 x 10000 = 1,000,000 Ω or 1MΩ
- R102 = 0.102 Ω (4-digit SMD resistors (E96 series)
- 0R10 = 0.1 x 100 = 0.1 x 1 = 0.1 Ω (4-digit SMD resistors (E24 series)
- 25R5 = 25.5Ω (4-digit SMD resistors (E96 series))

#### Capacitor part numbers

Most capacitors values are encoded in a 3-digit number where the 1st two digits
are the value and the last digit is the number of zeros in pF. Since we have 4
digits, the 1st digit is typically not used, but can if precision caps exist
that need 3 significant places to encode the value. The goal is to match what
most vendors are doing so we can easily compare IPN and vendor part numbers.

Examples:

- 103 = 10 \* 10^3 = 10,000pF = 10nF = 0.01uF
- 104 = 0.1uF

To figure out the extension, you can divide the capitance by 1pF to get the
number of pF. From this, you can visually tell what the variation should be.
Example:

`0.022uF = 0.022e-6/1e-12 = 22000`

So the extension would be `223`

To work backwards, we would have `1000 * 22 = 22,000pF/1e-6 = 0.022uF`.

## Implementation

Defining a part number structure is only part of the story -- implementation is
also critical. IPNs function as a _common_ reference to an object across an
organization. Thus, the implementation needs to be common across the
organization. Engineers should be able to pull new PN's and specify requirements
in the same database as manufacturing uses for planning and purchasing -- this
is the only configuration that will scale.

## Structured or Unstructured?

One of the fundamental questions regarding part numbers is whether to use
structured or unstructured part numbers. An unstructured part number is a number
that starts at say 1000000, and simply increments for each new part.

A fully structured part number might try to encode every parameter in the part
number -- for example:

`RES-0603-0.1W-1%-20ppm-10K`

A semi-structured part number might be:

`RES-025-1002`

The `0603-0.1W-1%-20ppm` parameters are all represented by the NNN section
(`025`). `1002` is the standard EIA E69 coding for 10K, which is used to encode
the value in most 1% resistor manufacturer part numbers today.

Many companies use a semi-structured part number format that consists of the
following components:

- **category** - a broad category for the part
- **incrementing number** - this is a simple incrementing number within a
  category that gives semi-structured part numbers all the same flexibility that
  unstructured part numbers have.
- **variation** - this field is used to differentiate similar parts and can
  encode the differentiating parameter(s) (resistance, length, size, etc) or can
  be a simple incrementing number.

There are many arguments for and against structure in part numbers, and
different organizations have different needs, so there is no one-size-fits-all.
Some trade-offs to consider:

- **semi-structure cons**
  - some training and knowledge is required to use the system
  - some claim numbers (vs letters) are easier to type on a keypad. However,
    more and more people are using laptops which don't have a numberpad.
  - the business may change such that the structure you start with no longer
    makes sense
  - parts may be categorized incorrectly, which is hard to fix later
- **semi-structure pros**
  - a semi-structured part number like CAP-023-0429 is easy to recognize as an
    IPN, and differentiate from other numbers
  - the category (CCC) part in easy to recognize/remember which reduces the size
    of the arbitrary NNN section you need to memorize. This also naturally sorts
    parts for you -- on your BOM, in the warehouse, in the factory, in your lab
    stock, etc. When physically dealing with 100's part numbers on a large PCB,
    any organization is helpful.
  - the NNN increments, so you still have a "Random" part, which gives you any
    flexibility you need.
  - the VVVV allows you to group variations of the same part together -- in
    documentation and in the warehouse.
  - RES-098-1004 is much easier for a human to compare without mistakes
    than 1029102.
  - phone numbers, two-factor authentication codes, PINs, etc are often in the
    form of XXX-XXX. There is a reason for this -- groups of 3 letters/digits
    are easy for humans to remember and recognize. Why not follow a similar
    format in IPNs?
  - We often group schematic parts into symbol libraries in our CAD tools. If
    that level of organization is useful for designers, why not leverage that
    same organization throughout the company?
  - humans read and compare IPNs many more times than we create or write them.
    Thus IPNs **should be optimized for reading and comparison by humans.**
  - a semi-structured part number is conducive to simple automation and
    scripting tasks. You can do a lot with a few lines of code (GitPLM is an
    example of this). CCC values can trigger different types of workflow. This
    helps a lot when you are starting out and can't afford $2M for a full blown
    MRP/PLM/... system.
  - a semi-structured part number is simple to implement -- any company, no
    matter how small, can implement this now and gain benefits.

An argument can be made that we don't need semi-structured IPNs and a simple
incrementing number and expressive descriptions/parameters in a database is
adequate. This works fine if you are using a computer and have access to a
database. However, this does not help you when you are in the warehouse picking
parts and comparing bags in a bin to numbers on a BOM, or doing a quick scan of
a BOM looking for mistakes. In this scenario, having **easy to compare
identifiers** helps a lot. This **improves efficiency** and **minimizes
mistakes**.

The case can be made that inventory and PLM software does care about structure
in PNs. This may be true, but you still need to solve the following problems:

- humans move parts around -- smart naming helps minimize errors and improve
  processing/recognition/comparison
- how are you going to organize/find parts/products in your back room before you
  are big enough to justify an inventory management system? If you have a
  structured PN system, you have a built-in way to sort and find parts.
- you still have to organize parts when kitting them for moving around, staging
  for manufacturing operations, etc.

An argument can be made that if you can't encode all parameters in the PN, then
you should not encode any. Information is not all or nothing. Some information
is better than no information. CCC/VVVV organization is very useful when dealing
with physical bags of parts in the real world. To find commonly used parts, the
CCC code is fairly obvious -- you will quickly memorize the small NNN number for
commonly used parts (like 1%, 0603 resistors) and the variation can often be
deduced logically for most parts. The reason we use IPNs is similar to why we
give people, countries, cities, etc. names -- so we can quickly identify and
communicate information about something in the physical (or even virtual) world
between humans. We don't call our co-worker down the hall
`engineer-5'11"-brown-hair-blue-eyes-150lbs-bsee-...` or `1029629` -- we
identify people by `<firstname> <lastname>`. Names are useful!

The CCC-NNN-VVVV scheme is a pragmatic compromise between random part numbers
and extensive descriptions where every last parameter is encoded in the
description. It is kind of like using colors on the factory floor -- you can't
encode everything in colors, but what you can encode sure helps with rapid,
accurate processing by humans.

Three letters for CCC has the following attributes:

- descriptive enough that you can encode meaning in it -- RES, CAP, SWI, DIO,
  etc are all fairly obvious and easier to remember/recognize than an arbitrary
  number.
- short enough to naturally limit the number of categories. If you had 4
  characters, you'd probably end up with RESS (surface mount resistor), and REST
  (through hole resistor), which is probably overkill and just complicates part
  number assignment.

CCC-NNN-VVVV also follows the general to specific naming convention, which is
generally a good way to name things.

If you are worried about running out of CCC/NNN variations, then make it
CCCC-NNNN-VVVV. Still very easy to manually parse on the floor. However, the
original NNN size gives you 1000 values, which is a lot. If your company is
successful to the point that you need more than 1000 values in a category, then
it's time to celebrate, and you'll certainly have the resources to handle the
next phase of expanding it to NNNN, more categories, etc. However, **you first
need to get there, so optimize for the "getting to success" phase now.** All the
microseconds and mistakes you'll save with the shorter NNN vs the longer NNNN
will pay off. And there is no reason why NNN and NNNN can't live in the same
database post success -- you now have an additional piece of information which
groups parts in time.

The CCC-NNN-VVVV format presented here is optimized for a small/mid-sized
company making electronic products. It may not be optimal for other industries.

If you do a google search on this topic, the seemingly prevailing opinion is
against structured part numbers. However, it appears most of these articles are
written by PLM tool vendors. Perhaps their criticisms are valid for fully
structured part numbers, but we're already demonstrated that semi-structured
part numbers can be designed to avoid most drawbacks, other than a little more
work up-front to create. The efficiencies gained downstream should pay back this
effort many times. It appears that most automotive manufacturers use
semi-structured part numbers. I don't have direct experience, but I've heard you
can tell where a part goes on an automobile from its PN. It takes more work up
front to figure out these part numbers, but having this bit of information
during manufacturing is a simple and effective check against errors and improves
efficiency. McMaster-Carr also uses semi-structured PNs. Maybe you should too
...

## Reference

The above information was compiled from the following discussions, articles, and
direct discussions with various people. All input, especially criticisms, has
been very valuable in clarifying the thinking on this topic.

- [Extensive discussion on the KiCad forum](https://forum.kicad.info/t/internal-house-part-number-formats/34958)
- PLM good practice (PDexpert)
  - [Part numbering system design](https://www.buyplm.com/plm-good-practice/part-numbering-system-software.aspx)
  - [Intelligent part numbers: The cost of being too smart](https://www.buyplm.com/plm-good-practice/intelligent-part-number-scheme.aspx)
- [Intelligent Numbering: What’s the Great Part Number Debate?](https://blog.grabcad.com/blog/2014/07/24/intelligent-numbering-debate/)
  - 4 perspectives to consider:
    - Creation and Data Entry
    - Longevity and Legacies
    - Readability
    - Uniqueness
    - Interpretation
  - This is a great article that explains the challenge of balancing all these
    factors.
- Oleg Shilovitsky
  - [Why to use intelligent PNs in the 21st century](https://beyondplm.com/2015/09/18/why-to-use-intelligent-part-numbers-in-21st-century/)
  - [Part Numbers are hard. How to think about data first?](https://beyondplm.com/2014/07/28/part-numbers-are-hard-how-to-think-about-data-first/)
    - > Product data is one of the most expensive assets in manufacturing
      > companies. It represents your company IP and it is a real foundation of
      > every manufacturing business. Think about data first. It will help you
      > to develop strategy that organize data for longer lifecycle and minimize
      > the cost of bringing new systems and manage changes in existing systems.
  - [The future of Part Numbers and Unique Identification?](https://beyondplm.com/2013/12/12/the-future-of-part-numbers-and-unique-identification/)
    - discusses schemes that large companies are using to accomplish global
      product IDs
